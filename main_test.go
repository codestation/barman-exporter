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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

var fakeExitCode = 0

type fakeClock struct{}

func (fakeClock) Now() time.Time                         { return time.Date(2022, 2, 3, 12, 15, 0, 0, time.UTC) }
func (fakeClock) After(d time.Duration) <-chan time.Time { return time.After(0) }

func TestAll(t *testing.T) {
	execCommand = fakeExecCommand
	clock = fakeClock{}
	assert.NoError(t, collectMetrics())
	testBarmanData, _ := os.Open("tests/metrics_test.txt")
	assert.NoError(t, testutil.GatherAndCompare(prometheus.DefaultGatherer, testBarmanData,
		"barman_status",
		"barman_last_wal_age_seconds",
		"barman_last_backup_age_seconds",
		"barman_last_backup_size_bytes",
		"barman_backup_duration_seconds",
		"barman_backup_window_seconds",
	))
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "GO_FAKE_EXIT_CODE=" + strconv.Itoa(fakeExitCode)}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	command := os.Args[3]
	arguments := os.Args[6:]

	switch command {
	case "barman":
		switch arguments[0] {
		case "status":
			jsonFile, err := ioutil.ReadFile("tests/status_test.json")
			if err != nil {
				panic(err.Error())
			}
			_, _ = fmt.Fprint(os.Stdout, string(jsonFile))
		case "check":
			jsonFile, err := ioutil.ReadFile("tests/check_test.json")
			if err != nil {
				panic(err.Error())
			}
			_, _ = fmt.Fprint(os.Stdout, string(jsonFile))
		case "list-server":
			jsonFile, err := ioutil.ReadFile("tests/list_server_test.json")
			if err != nil {
				panic(err.Error())
			}
			_, _ = fmt.Fprint(os.Stdout, string(jsonFile))
		case "list-backup":
			jsonFile, err := ioutil.ReadFile("tests/list_backup_test.json")
			if err != nil {
				panic(err.Error())
			}
			_, _ = fmt.Fprint(os.Stdout, string(jsonFile))
		case "show-backup":
			jsonFile, err := ioutil.ReadFile("tests/" + arguments[2] + "_show_backup_test.json")
			if err != nil {
				panic(err.Error())
			}
			_, _ = fmt.Fprint(os.Stdout, string(jsonFile))
		default:
			_, _ = fmt.Fprintf(os.Stderr, "Unknown barman command call")
			os.Exit(1)
		}
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Unknown command call: %s", command)
		os.Exit(2)
	}

	os.Exit(0)
}
