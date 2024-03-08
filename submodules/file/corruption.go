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
	"path/filepath"
	"strconv"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = corruption{
		FaultType: "file-corruption",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type corruption struct {
	FaultType              string
	flags                  map[string]string
	filePath               string
	size                   int64
	backupPath             string
	backupFilePath         string
	backupPathMntAvailSize int64
	offset                 int64
	length                 int64
}

func (c *corruption) setBackupPathMntAvailSize() error {
	var backupPathMntAvailSize int64
	// 判断是否自定义备份路径，如果自定备份路径，文件备份在对应路径，
	// 如果没有自定义路径，备份路径与原文件在相同的文件夹下。
	if c.backupPath != "" {
		fileInfo, err := os.Stat(c.backupPath)
		if err != nil {
			return fmt.Errorf("backup path(%s) not exist", c.backupPath)
		}
		if !fileInfo.IsDir() {
			return fmt.Errorf("please input valid backup directory")
		}

		backupPathMntAvailSize, err = getPathMountPointAvailSize(c.backupPath)
		if err != nil {
			return fmt.Errorf("get file backup path avail size failed(%v)", err)
		}
	} else {
		var err error
		backupPathMntAvailSize, err = getPathMountPointAvailSize(c.filePath)
		if err != nil {
			return fmt.Errorf("get file backup path avail size failed(%v)", err)
		}
	}
	c.backupPathMntAvailSize = backupPathMntAvailSize
	return nil
}

func (c *corruption) setBackupFilePath() error {
	if backupPath, ok := c.flags["backup-path"]; ok {
		_, fileName := filepath.Split(c.filePath)
		c.backupFilePath = fmt.Sprintf("%s/%s-%s-backup", backupPath, fileName, c.FaultType)
	} else {
		c.backupFilePath = fmt.Sprintf("%s-%s-backup", c.filePath, c.FaultType)
	}
	return nil
}

func (c *corruption) setOffsetAndLength() error {
	offsetStr, ok := c.flags["offset"]
	if !ok {
		return fmt.Errorf("please input file corruption offset")
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		return fmt.Errorf("trans offset(%s) to int64 failed(%v)", offsetStr, err)
	}
	c.offset = offset

	lengthStr, ok := c.flags["length"]
	if !ok {
		return fmt.Errorf("please input file corruption length")
	}
	length, err := strconv.ParseInt(lengthStr, 10, 64)
	if err != nil {
		return fmt.Errorf("trans length(%s) to int64 failed(%v)", lengthStr, err)
	}
	c.length = length
	return nil
}

func (c *corruption) Prepare(inputArgs []string) error {
	c.flags = parse.TransInputFlagsToMap(inputArgs)
	if backupPath, ok := c.flags["backup-path"]; ok {
		c.backupPath = backupPath
	}

	filePath, err := getFilePathAndCheck(c.flags)
	if err != nil {
		return fmt.Errorf("%s get full path or check error", c.FaultType)
	}
	c.filePath = filePath

	if err := c.setBackupPathMntAvailSize(); err != nil {
		return err
	}
	if err := c.setBackupFilePath(); err != nil {
		return err
	}
	return c.setOffsetAndLength()
}

func (c *corruption) FaultInject(_ []string) error {
	var err error
	c.size, err = getFileSize(c.filePath)
	if err != nil {
		return fmt.Errorf("fault type(%s) get file size failed(%v)", c.FaultType, err)
	}

	// 如果文件size大于备份挂载点avail size将会导致备份失败。
	if c.size > c.backupPathMntAvailSize {
		return fmt.Errorf("backup path(%s) not has enough space", filepath.Dir(c.backupFilePath))
	}

	// offset不能大于文件大小。
	if c.offset >= c.size {
		return fmt.Errorf("input offset(%d) >= file size(%d)", c.offset, c.size)
	}

	// offset+length不能大于文件大小。
	if c.length+c.offset >= c.size {
		return fmt.Errorf("input offset(%d) + length(%d) >= file size(%d)", c.offset, c.length, c.size)
	}

	// 备份文件。
	if _, err = fileCopy(c.filePath, c.backupFilePath); err != nil {
		return fmt.Errorf("backup file %s to %s failed, Err: %s", c.filePath, c.backupFilePath, err)
	}

	// 往文件offset处写长度为length的随机字符串。
	file, err := os.OpenFile(c.filePath, os.O_WRONLY, openFilePerm)
	if err != nil {
		return fmt.Errorf("open corruption file(%s) target failed(%v)", c.filePath, err)
	}
	defer file.Close()

	if _, err = file.Seek(c.offset, 0); err != nil {
		return fmt.Errorf("seek corruption file(%s) target failed(%v)", c.filePath, err)
	}

	randStr, err := getRandomString(c.length)
	if err != nil {
		return fmt.Errorf("%s get random string failed: %v", c.FaultType, err)
	}
	if _, err = file.WriteString(randStr); err != nil {
		return fmt.Errorf("make file(%s) corruption target failed(%v)", c.filePath, err)
	}
	return nil
}

func (c *corruption) FaultRemove(_ []string) error {
	if !util.FileIsExist(c.backupFilePath) {
		return fmt.Errorf("not found backup file path(%s)", c.backupFilePath)
	}

	if err := os.Remove(c.filePath); err != nil {
		return fmt.Errorf("remove corruption file(%s) failed(%v)", c.filePath, err)
	}

	mvShellCmd := fmt.Sprintf("mv %s %s", c.backupFilePath, c.filePath)
	if _, err := util.ExecCommandBlock(mvShellCmd); err != nil {
		return fmt.Errorf("restore corruption file(%s) failed(%v)", c.filePath, err)
	}
	return nil
}
