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
	"encoding/json"
	"os/exec"
)

var execCommand = exec.Command
var barmanPath = "barman"

func barmanCheck(server string) (BarmanCheck, error) {
	cmd := execCommand(barmanPath, "-f", "json", "check", server)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	data := BarmanCheck{}
	if err = json.Unmarshal(output, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func barmanListServer() (BarmanListServer, error) {
	cmd := execCommand(barmanPath, "-f", "json", "list-server")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	data := BarmanListServer{}
	if err = json.Unmarshal(output, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func barmanListBackup(server string) (BarmanListBackup, error) {
	cmd := execCommand(barmanPath, "-f", "json", "list-backup", server)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	data := BarmanListBackup{}
	if err = json.Unmarshal(output, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func barmanStatus(server string) (BarmanStatus, error) {
	cmd := execCommand(barmanPath, "-f", "json", "status", server)
	result, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var data BarmanStatus
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func barmanShowBackup(server, id string) (BarmanShowBackup, error) {
	cmd := execCommand(barmanPath, "-f", "json", "show-backup", server, id)
	result, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var data BarmanShowBackup
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, err
	}

	return data, nil
}
