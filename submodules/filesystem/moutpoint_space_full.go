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
	"math"
	"os"
	"strings"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = moutpointSpaceFull{
		FaultType: "filesystem-mountpoint-space-full",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type moutpointSpaceFull struct {
	FaultType  string
	flags      map[string]string
	size       float64
	imgPath    string
	mountPoint string
}

func (m *moutpointSpaceFull) Prepare(inputArgs []string) error {
	dependCmd := []string{"df", "dd"}
	if missingCmd, isMissCmd := util.CheckEnvShellCommand(dependCmd); isMissCmd {
		return fmt.Errorf("missing command: %s", missingCmd)
	}

	m.flags = parse.TransInputFlagsToMap(inputArgs)
	mountPoint, err := mountPointCheck(m.flags)
	if err != nil {
		return fmt.Errorf("mountPoint check failed: %v", err)
	}
	m.mountPoint = mountPoint

	size, err := getMountPointSize(mountPoint)
	if err != nil {
		return fmt.Errorf("get mount point size failed, Error: %s", err)
	}
	m.size = size

	// dd命令消耗挂载点磁盘空间的文件命名为$mountPoint/$faultType-image。
	m.imgPath = fmt.Sprintf("%s/%s-image", mountPoint, m.FaultType)
	return nil
}

func (m *moutpointSpaceFull) FaultInject(_ []string) error {
	// TODO: 剩余可用磁盘空间可能大于某个特定文件系统支持单个文件的最大size。
	if util.FileIsExist(m.imgPath) {
		return fmt.Errorf("path: %s has been injected: %s fault", m.mountPoint, m.FaultType)
	}

	// TODO: bs大小是否可以动态获取读取效率最高值。
	ddCmd := fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=%d > /dev/null 2>&1 &",
		m.imgPath, int(math.Ceil(m.size)))
	if result, err := util.ExecCommandBlock(ddCmd); err != nil {
		return fmt.Errorf("execute: %s error: %v, result: %s", ddCmd, err, result)
	}
	return nil
}

func (m *moutpointSpaceFull) killBackgroundInjectProcess() error {
	searchStr := fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=%d", m.imgPath, int(math.Ceil(m.size)))
	getPidShellCmd := fmt.Sprintf("ps aux | grep -v grep | grep -w '%s' | awk '{print $2}'", searchStr)
	pidStr, err := util.ExecCommandBlock(getPidShellCmd)
	if err != nil {
		return fmt.Errorf("failed to obtain pid of dd process running in the background")
	}
	if pidStr == "" {
		return nil
	}

	killCmd := fmt.Sprintf("kill -9 %s", strings.ReplaceAll(pidStr, "\n", " "))
	if result, err := util.ExecCommandBlock(killCmd); err != nil {
		return fmt.Errorf("execute command: %s failed, err: %v result: %s", killCmd, err, result)
	}
	return nil
}

func (m *moutpointSpaceFull) FaultRemove(_ []string) error {
	// 存在用户提前做清理的场景，需要将dd后台执行进程kill掉。
	if err := m.killBackgroundInjectProcess(); err != nil {
		return fmt.Errorf("%s kill background inject process failed(%v)", m.FaultType, err)
	}

	if isExist := util.FileIsExist(m.imgPath); isExist {
		if err := os.Remove(m.imgPath); err != nil {
			return fmt.Errorf("remove file: %s error: %v", m.imgPath, err)
		}
	}
	return nil
}
