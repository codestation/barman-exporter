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

import "encoding/json"

type HintStatus struct {
	Hint   string `json:"hint"`
	Status string `json:"status"`
}

type CheckInfo struct {
	ArchiveCommand                              HintStatus `json:"archive_command"`
	ArchiveMode                                 HintStatus `json:"archive_mode"`
	ArchiverErrors                              HintStatus `json:"archiver_errors"`
	BackupMaximumAge                            HintStatus `json:"backup_maximum_age"`
	BackupMinimumSize                           HintStatus `json:"backup_minimum_size"`
	CompressionSettings                         HintStatus `json:"compression_settings"`
	ContinuousArchiving                         HintStatus `json:"continuous_archiving"`
	Directories                                 HintStatus `json:"directories"`
	FailedBackups                               HintStatus `json:"failed_backups"`
	MinimumRedundancyRequirements               HintStatus `json:"minimum_redundancy_requirements"`
	PgReceivexlog                               HintStatus `json:"pg_receivexlog"`
	PgReceivexlogCompatible                     HintStatus `json:"pg_receivexlog_compatible"`
	Postgresql                                  HintStatus `json:"postgresql"`
	PostgresqlStreaming                         HintStatus `json:"postgresql_streaming"`
	ReceiveWalRunning                           HintStatus `json:"receive_wal_running"`
	ReplicationSlot                             HintStatus `json:"replication_slot"`
	RetentionPolicySettings                     HintStatus `json:"retention_policy_settings"`
	SSH                                         HintStatus `json:"ssh"`
	SuperuserOrStandardUserWithBackupPrivileges HintStatus `json:"superuser_or_standard_user_with_backup_privileges"`
	SystemidCoherence                           HintStatus `json:"systemid_coherence"`
	WalLevel                                    HintStatus `json:"wal_level"`
	WalMaximumAge                               HintStatus `json:"wal_maximum_age"`
	WalSize                                     HintStatus `json:"wal_size"`
}

func (c CheckInfo) AllOk() bool {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return false
	}
	var fields map[string]HintStatus
	if err = json.Unmarshal(jsonData, &fields); err != nil {
		return false
	}
	for field := range fields {
		if fields[field].Status != "OK" {
			return false
		}
	}

	return true
}

type BarmanCheck map[string]CheckInfo

type DescriptionMessage struct {
	Description string `json:"description"`
	Message     string `json:"message"`
}

type StatusInfo struct {
	Active                    DescriptionMessage `json:"active"`
	ArchiveCommand            DescriptionMessage `json:"archive_command"`
	BackupsNumber             DescriptionMessage `json:"backups_number"`
	CurrentSize               DescriptionMessage `json:"current_size"`
	CurrentXlog               DescriptionMessage `json:"current_xlog"`
	DataDirectory             DescriptionMessage `json:"data_directory"`
	Description               DescriptionMessage `json:"description"`
	Disabled                  DescriptionMessage `json:"disabled"`
	FailedCount               DescriptionMessage `json:"failed_count"`
	FirstBackup               DescriptionMessage `json:"first_backup"`
	IsInRecovery              DescriptionMessage `json:"is_in_recovery"`
	LastArchivedWal           DescriptionMessage `json:"last_archived_wal"`
	LastBackup                DescriptionMessage `json:"last_backup"`
	MinimumRedundancy         DescriptionMessage `json:"minimum_redundancy"`
	PassiveNode               DescriptionMessage `json:"passive_node"`
	PgVersion                 DescriptionMessage `json:"pg_version"`
	Pgespresso                DescriptionMessage `json:"pgespresso"`
	RetentionPolicies         DescriptionMessage `json:"retention_policies"`
	ServerArchivedWalsPerHour DescriptionMessage `json:"server_archived_wals_per_hour"`
}

type BarmanStatus map[string]StatusInfo

type ListInfo struct {
	Description string `json:"description"`
}
type BarmanListServer map[string]ListInfo

type BaseBackupInformation struct {
	AnalysisTime           string  `json:"analysis_time"`
	AnalysisTimeSeconds    float64 `json:"analysis_time_seconds"`
	BeginLsn               string  `json:"begin_lsn"`
	BeginOffset            int     `json:"begin_offset"`
	BeginTime              string  `json:"begin_time"`
	BeginTimeTimestamp     string  `json:"begin_time_timestamp"`
	BeginWal               string  `json:"begin_wal"`
	CopyTime               string  `json:"copy_time"`
	CopyTimeSeconds        float64 `json:"copy_time_seconds"`
	DiskUsage              string  `json:"disk_usage"`
	DiskUsageBytes         int64   `json:"disk_usage_bytes"`
	DiskUsageWithWals      string  `json:"disk_usage_with_wals"`
	DiskUsageWithWalsBytes int64   `json:"disk_usage_with_wals_bytes"`
	EndLsn                 string  `json:"end_lsn"`
	EndOffset              int     `json:"end_offset"`
	EndTime                string  `json:"end_time"`
	EndTimeTimestamp       string  `json:"end_time_timestamp"`
	EndWal                 string  `json:"end_wal"`
	IncrementalSize        string  `json:"incremental_size"`
	IncrementalSizeBytes   int64   `json:"incremental_size_bytes"`
	IncrementalSizeRatio   string  `json:"incremental_size_ratio"`
	NumberOfWorkers        int     `json:"number_of_workers"`
	Throughput             string  `json:"throughput"`
	ThroughputBytes        float64 `json:"throughput_bytes"`
	Timeline               int     `json:"timeline"`
	WalCompressionRatio    string  `json:"wal_compression_ratio"`
}

type CatalogInformation struct {
	NextBackup      string `json:"next_backup"`
	PreviousBackup  string `json:"previous_backup"`
	RetentionPolicy string `json:"retention_policy"`
}

type WalInformation struct {
	CompressionRatio string        `json:"compression_ratio"`
	DiskUsage        string        `json:"disk_usage"`
	DiskUsageBytes   int           `json:"disk_usage_bytes"`
	LastAvailable    string        `json:"last_available"`
	NoOfFiles        int           `json:"no_of_files"`
	Timelines        []interface{} `json:"timelines"`
	WalRate          string        `json:"wal_rate"`
	WalRatePerSecond float64       `json:"wal_rate_per_second"`
}

type ShowBackupInfo struct {
	BackupID              string `json:"backup_id"`
	BaseBackupInformation `json:"base_backup_information"`
	CatalogInformation    CatalogInformation `json:"catalog_information"`
	PgdataDirectory       string             `json:"pgdata_directory"`
	PostgresqlVersion     int                `json:"postgresql_version"`
	Status                string             `json:"status"`
	Tablespaces           []interface{}      `json:"tablespaces"`
	WalInformation        WalInformation     `json:"wal_information"`
}

type BarmanShowBackup map[string]ShowBackupInfo

type BackupInfo struct {
	BackupID         string        `json:"backup_id"`
	EndTime          string        `json:"end_time"`
	EndTimeTimestamp string        `json:"end_time_timestamp"`
	RetentionStatus  string        `json:"retention_status"`
	Size             string        `json:"size"`
	SizeBytes        int64         `json:"size_bytes"`
	Status           string        `json:"status"`
	Tablespaces      []interface{} `json:"tablespaces"`
	WalSize          string        `json:"wal_size"`
	WalSizeBytes     int           `json:"wal_size_bytes"`
}

type BarmanListBackup map[string][]BackupInfo
