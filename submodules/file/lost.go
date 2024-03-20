/*
Copyright 2023 Sangfor Technologies Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file

import (
	"fmt"
	"os"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
)

func init() {
	var newFaultType = lost{
		FaultType: "file-lost",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type lost struct {
	FaultType      string
	flags          map[string]string
	filePath       string
	backupFilePath string
}

func (f *lost) Prepare(inputArgs []string) error {
	f.flags = parse.TransInputFlagsToMap(inputArgs)
	inputFilePath, ok := f.flags["path"]
	if !ok {
		return fmt.Errorf("%s please input file path", f.FaultType)
	}

	filePath, err := getFileFullPath(inputFilePath)
	if err != nil {
		return fmt.Errorf("get file %s full path failed, error: %s", filePath, err)
	}
	f.filePath = filePath

	// 备份文件名命名为$(filename)-$(faultType)-backup。
	f.backupFilePath = fmt.Sprintf("%s-%s-backup", f.filePath, f.FaultType)
	return nil
}

func (f *lost) FaultInject(_ []string) error {
	if _, err := os.Stat(f.filePath); err != nil {
		return fmt.Errorf("please check file %s exist %v", f.filePath, err)
	}

	if err := os.Rename(f.filePath, f.backupFilePath); err != nil {
		return fmt.Errorf("file %s injection %s fault failed, Error: %w", f.filePath, f.FaultType, err)
	}
	return nil
}

func (f *lost) FaultRemove(_ []string) error {
	if _, err := os.Stat(f.backupFilePath); err != nil {
		return fmt.Errorf("please check backup file %s exist", f.backupFilePath)
	}

	if err := os.Rename(f.backupFilePath, f.filePath); err != nil {
		return fmt.Errorf("file %s clearing %s fault failed, error: %w", f.filePath, f.FaultType, err)
	}
	return nil
}
