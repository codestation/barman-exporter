/*
 *
 * Copyright 2022 codestation.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
)

const versionFormatter = `barman-exporter version: %s, commit: %s, built at: %s`

var (
	status = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "barman_status",
		Help: "1 if server passes all diagnostics",
	}, []string{"server"})
	lastWalAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "barman_last_wal_age_seconds",
		Help: "Time since last received wal",
	}, []string{"server"})
	lastBackupAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "barman_last_backup_age_seconds",
		Help: "Time since last full backup",
	}, []string{"server"})
	lastBackupSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "barman_last_backup_size_bytes",
		Help: "Size of last backup",
	}, []string{"server"})
	backupDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "barman_backup_duration_seconds",
		Help: "Duration of last backup",
	}, []string{"server"})
	backupWindow = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "barman_backup_window_seconds",
		Help: "Time range for PITR",
	}, []string{"server"})
)

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

var clock Clock = realClock{}

func printVersion(c *cli.Context) {
	_, _ = fmt.Fprintf(c.App.Writer, versionFormatter, Version, Commit, BuildTime)
}

func addGaugeServer(gauge *prometheus.GaugeVec, server string) prometheus.Gauge {
	return gauge.With(prometheus.Labels{"server": server})
}

func convertDateToTimestamp(date string) int64 {
	const ctLayout = "Mon Jan 2 15:04:05 2006"
	dateTime, err := time.Parse(ctLayout, date)
	if err != nil {
		return -1
	} else {
		return dateTime.Unix()
	}
}

func collectMetrics() error {
	servers, err := barmanListServer()
	if err != nil {
		log.Printf("failed to run barman list-server: %v", err)
	}

	for server := range servers {
		serverCheck, err := barmanCheck(server)
		if err == nil {
			check := serverCheck[server]
			if check.AllOk() {
				addGaugeServer(status, server).Set(1)
			} else {
				addGaugeServer(status, server).Set(0)
			}
		} else {
			log.Printf("Failed to run barman check %s: %v", server, err)
		}

		now := clock.Now()
		var lastWalTimestamp int64

		infoList, err := barmanStatus(server)
		if err == nil {
			info := infoList[server]
			dateParts := strings.Split(info.LastArchivedWal.Message, ", at ")
			if len(dateParts) == 2 {
				lastWalTimestamp = convertDateToTimestamp(dateParts[1])
				addGaugeServer(lastWalAge, server).Set(float64(now.Unix() - lastWalTimestamp))
			}

		} else {
			log.Printf("Failed to run barman status %s: %v", server, err)
		}

		backups, err := barmanListBackup(server)
		if err != nil {
			log.Printf("Failed to run barman list-backup %s: %v", server, err)
		} else {
			backupList := backups[server]
			var backupEntries []BackupInfo
			for _, entry := range backupList {
				if entry.Status == "DONE" {
					backupEntries = append(backupEntries, entry)
				}
			}

			if len(backupEntries) > 0 {
				first := backupEntries[len(backupEntries)-1]
				last := backupEntries[0]
				addGaugeServer(lastBackupSize, server).Set(float64(last.SizeBytes))
				showList, err := barmanShowBackup(server, last.BackupID)
				if err != nil {
					log.Printf("Failed to run barman show-backup %s %s: %v", server, last.BackupID, err)
				}

				showLast := showList[server]
				backupStart, err := strconv.ParseInt(showLast.BeginTimeTimestamp, 10, 64)
				if err == nil {
					addGaugeServer(lastBackupAge, server).Set(float64(now.Unix() - backupStart))
				} else {
					log.Printf("failed to convert BeginTime timestamp: %v", err)
				}

				showList, err = barmanShowBackup(server, first.BackupID)
				if err != nil {
					log.Printf("Failed to run barman show-backup %s %s: %v", server, last.BackupID, err)
				}

				showFirst := showList[server]
				backupEnd, err := strconv.ParseInt(showLast.EndTimeTimestamp, 10, 64)
				if err == nil {
					addGaugeServer(backupDuration, server).Set(float64(backupEnd - backupStart))
				} else {
					log.Printf("failed to convert EndTime timestamp: %v", err)
				}

				firstFull, err := strconv.ParseInt(showFirst.BeginTimeTimestamp, 10, 64)
				if err == nil {
					addGaugeServer(backupWindow, server).Set(float64(lastWalTimestamp - firstFull))
				}
			}
		}
	}

	return nil
}

func collectMetricsLoop(ctx context.Context, signal chan os.Signal, interval time.Duration) error {
	if err := collectMetrics(); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			log.Printf("Exiting metrics loop")
			return nil // avoid leaking of this goroutine when ctx is done.
		case <-signal:
			log.Printf("Running metrics (SIGUSR1)")
			if err := collectMetrics(); err != nil {
				return err
			}
		case <-time.After(interval):
			log.Printf("Running metrics")
			if err := collectMetrics(); err != nil {
				return err
			}
		}
	}
}

func run(c *cli.Context) error {
	if c.IsSet("barman-path") {
		barmanPath = c.String("barman-path")
	}

	c1, cancel := context.WithCancel(context.Background())
	s := http.Server{Addr: c.String("listen")}

	signalUsr := make(chan os.Signal, 1)
	signal.Notify(signalUsr, syscall.SIGUSR1)
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

	go func(signal chan os.Signal) {
		if err := collectMetricsLoop(c1, signal, c.Duration("interval")); err != nil {
			log.Printf("failed to collect metrics: %v", err)
		}
		exitCh <- os.Interrupt
	}(signalUsr)

	r := prometheus.NewRegistry()
	r.MustRegister(status, lastWalAge, lastBackupAge, lastBackupSize, backupDuration, backupWindow)
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

	http.Handle(c.String("metrics-path"), handler)
	log.Printf("Starting web server")

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Print(err)
		}
	}()

	log.Printf("Waiting for metrics loop to finish")
	<-exitCh

	log.Printf("Stopping web server")
	if err := s.Shutdown(c1); err != nil {
		log.Print(err)
	}

	cancel()

	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "barman-exporter"
	app.Version = Version
	cli.VersionPrinter = printVersion

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "listen, l",
			Usage:   "listen address",
			Value:   ":8000",
			EnvVars: []string{"LISTEN"},
		},
		&cli.DurationFlag{
			Name:    "interval, i",
			Usage:   "interval",
			Value:   time.Minute * 5,
			EnvVars: []string{"INTERVAL"},
		},
		&cli.StringFlag{
			Name:    "metrics-path, m",
			Usage:   "metrics path",
			Value:   "/metrics",
			EnvVars: []string{"METRICS_PATH"},
		},
		&cli.StringFlag{
			Name:    "barman-path, p",
			Usage:   "barman path",
			Value:   "barman",
			EnvVars: []string{"BARMAN_PATH"},
		},
	}

	app.Action = run

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Unrecoverable error: %s", err.Error())
	}
}
