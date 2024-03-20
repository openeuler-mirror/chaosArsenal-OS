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

package filesystem

import (
	"fmt"
	"strconv"
	"strings"

	"arsenal-os/util"

	"github.com/moby/sys/mountinfo"
)

const (
	unitIndex = 2
	kb        = 1 << 10
	bitSize   = 32
)

// getMountPointInfoByPath 通过传入路径返回挂载点信息。
func getMountPointInfoByPath(dirPath string) (*mountinfo.Info, error) {
	var recordInfo *mountinfo.Info
	_, err := mountinfo.GetMounts(func(info *mountinfo.Info) (bool, bool) {
		if info.Mountpoint == dirPath {
			recordInfo = info
			return false, true
		}
		return false, false
	})
	if err != nil {
		return nil, fmt.Errorf("retrieves a list of mounts for the current running process failed: %v", err)
	}
	return recordInfo, nil
}

func mountPointCheck(flags map[string]string) (string, error) {
	mountPoint, ok := flags["path"]
	if !ok {
		return "", fmt.Errorf("please make sure has been input mount-point parameter")
	}
	if !util.FileIsExist(mountPoint) {
		return "", fmt.Errorf("input mount point path: %s not exist", mountPoint)
	}

	mntInfo, err := getMountPointInfoByPath(mountPoint)
	if err != nil {
		return "", err
	}
	if mntInfo == nil {
		return "", fmt.Errorf("path: %s not a mount point", mountPoint)
	}
	// 如果挂载点只读，不允许注入文件系统挂载点相关故障。
	optionList := strings.Split(mntInfo.Options, ",")
	for _, option := range optionList {
		if option == "ro" {
			return "", fmt.Errorf("mount point: %s is read-only", mountPoint)
		}
	}
	return mntInfo.Mountpoint, nil
}

func transSizeToMb(inputValue string, unit byte) (float64, error) {
	var value float64
	rawValue, err := strconv.ParseFloat(inputValue, bitSize)
	if err != nil {
		return 0, err
	}

	unitStr := fmt.Sprintf("%c", unit)
	switch unitStr {
	case "M":
		value = rawValue
	case "G":
		value = rawValue * kb
	case "T":
		value = rawValue * kb * kb
	default:
		return 0, fmt.Errorf("please input the correct units")
	}
	return value, nil
}

// getMountPointSize 获取挂载点总磁盘空间，单位为MB。
func getMountPointSize(mntPoint string) (float64, error) {
	// 挂载点信息: "Filesystem Size Used Avail Use% Mounted on"。
	shellCmd := fmt.Sprintf("df -h %s | tail -n 1 | awk '{print $2}'", mntPoint)
	sizeInfo, err := util.ExecCommandBlock(shellCmd)
	if err != nil {
		return 0, fmt.Errorf("get mount point: %s size info failed: %s", mntPoint, err)
	}

	// 从命令行中返回的Size信息的末尾包含一个换行符 \n。
	if len(sizeInfo) <= unitIndex {
		return 0, fmt.Errorf("get mount point: %s size info failed", mntPoint)
	}
	unitIndex := len(sizeInfo) - unitIndex
	value, err := transSizeToMb(sizeInfo[:unitIndex], sizeInfo[unitIndex])
	if err != nil {
		return 0, fmt.Errorf("trans size to Mb failed: %s", err)
	}
	return value, nil
}
