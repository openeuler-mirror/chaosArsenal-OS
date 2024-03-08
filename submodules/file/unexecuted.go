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
	"io/ioutil"
	"os"
	"strconv"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = unexecuted{
		FaultType: "file-unexecuted",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type unexecuted struct {
	FaultType          string
	flags              map[string]string
	filePath           string
	fileMode           os.FileMode
	backupAttrFilePath string
	backupFileMode     os.FileMode
}

func (e *unexecuted) Prepare(inputArgs []string) error {
	e.flags = parse.TransInputFlagsToMap(inputArgs)
	filePath, err := getFilePathAndCheck(e.flags)
	if err != nil {
		return fmt.Errorf("%s get full path or check error", e.FaultType)
	}
	e.filePath = filePath

	fileInfo, err := os.Stat(e.filePath)
	if err != nil {
		return fmt.Errorf("get file(%s) info failed(%v)", e.filePath, err)
	}
	e.fileMode = fileInfo.Mode()
	e.backupAttrFilePath = fmt.Sprintf("%s-%s-backup-attr", filePath, e.FaultType)
	return nil
}

// fileRawAttributeBackup 备份文件原GUO属性信息。
func (e *unexecuted) fileRawAttributeBackup() error {
	file, err := os.OpenFile(e.backupAttrFilePath, os.O_CREATE|os.O_WRONLY, openFilePerm)
	if err != nil {
		return fmt.Errorf("create backup file(%s) attr file failed(%v)", e.filePath, err)
	}
	defer file.Close()

	mode := strconv.FormatInt(int64(uint32(e.fileMode.Perm())), 8)
	if _, err = file.WriteString(mode); err != nil {
		return fmt.Errorf("write file(%s) attr to %s failed(%v)", e.filePath, e.backupAttrFilePath, err)
	}
	return nil
}

func (e *unexecuted) FaultInject(_ []string) error {
	// 注入之前先判断文件是否真的已经具有可执行权限，只要有UGO一个Perm组中有执行权限认为对象合法。
	// --x--x--x 001001001 0x49
	const executeAttrMagic = 0x49
	if (e.fileMode & executeAttrMagic) == 0 {
		return fmt.Errorf("file(%v) no execute perm", e.filePath)
	}

	if err := e.fileRawAttributeBackup(); err != nil {
		return fmt.Errorf("backup up file attr failed(%v)", err)
	}

	// 将可以执行权限位置0 ---x--x--x -- -110110110。
	const executeAttrRevertMagic = 0x1b6
	if err := os.Chmod(e.filePath, e.fileMode&executeAttrRevertMagic); err != nil {
		return fmt.Errorf("file %s injection %s fault failed, Error: %s", e.filePath, e.FaultType, err)
	}
	return nil
}

func (e *unexecuted) setBackupFileAttribute() error {
	file, err := os.Open(e.backupAttrFilePath)
	if err != nil {
		return fmt.Errorf("backup attr file(%s) missing", e.backupAttrFilePath)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read backup attr file failed(%v)", err)
	}
	mode, err := strconv.ParseUint(string(data), 8, 32)
	if err != nil {
		return fmt.Errorf("trans file attr string to mode failed(%v)", err)
	}
	e.backupFileMode = os.FileMode(mode)
	return nil
}

func (e *unexecuted) FaultRemove(_ []string) error {
	if err := e.setBackupFileAttribute(); err != nil {
		return fmt.Errorf("file raw attr recover failed(%v)", err)
	}
	if err := os.Chmod(e.filePath, e.backupFileMode); err != nil {
		return fmt.Errorf("file %s clearing %s fault failed, Error: %s", e.filePath, e.FaultType, err)
	}

	if !util.FileIsExist(e.backupAttrFilePath) {
		return nil
	}
	if err := os.Remove(e.backupAttrFilePath); err != nil {
		return fmt.Errorf("remove backup attr file(%s) failed(%v)", e.backupAttrFilePath, err)
	}
	return nil
}
