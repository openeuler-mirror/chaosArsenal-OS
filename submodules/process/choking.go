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

package process

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = choking{
		FaultType: "process-choking",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type choking struct {
	FaultType string
	flags     map[string]string
	pid       int
	interval  int
}

// Prepare 获取输入参数maps，初始化opsInfo信息，检查进程是否存在。
func (c *choking) Prepare(inputArgs []string) error {
	dependCmd := []string{"ps", "grep", "awk"}
	if missingCmd, isMissCmd := util.CheckEnvShellCommand(dependCmd); isMissCmd {
		return fmt.Errorf("missing command: %s", missingCmd)
	}
	c.flags = parse.TransInputFlagsToMap(inputArgs)

	pid, err := GetProcessPidAndExistCheck(c.flags)
	if err != nil {
		return err
	}
	c.pid = pid

	if _, ok := c.flags["interval"]; !ok {
		return errors.New("prepare failed when get interval failed")
	}
	interval, err := strconv.Atoi(c.flags["interval"])
	if err != nil {
		return fmt.Errorf("prepare failed when trans interval string to int failed, error: %v", err)
	}
	c.interval = interval
	return nil
}

func (c *choking) FaultInject(_ []string) error {
	// TODO: 定时器实现。
	for {
		if err := syscall.Kill(c.pid, syscall.SIGSTOP); err != nil {
			return fmt.Errorf("send signal: SIGSTOP to %d failed", c.pid)
		}
		time.Sleep(time.Duration(c.interval) * time.Second)

		if err := syscall.Kill(c.pid, syscall.SIGCONT); err != nil {
			return fmt.Errorf("send signal: SIGCONT to %d failed", c.pid)
		}
		time.Sleep((time.Duration(c.interval)) * time.Second)
	}
}

func (c *choking) FaultRemove(inputArgs []string) error {
	// remove命令替换成inject命令并查找进程对应的pid。
	removeCommand := strings.Join(inputArgs, " ")
	injectCommand := strings.ReplaceAll(removeCommand, submodules.Remove, submodules.Inject)
	shellCmd := fmt.Sprintf("ps aux | grep '%s' | grep -v grep | awk '{print $2}'", injectCommand)
	pidStr, err := util.ExecCommandBlock(shellCmd)
	if err != nil {
		return fmt.Errorf("%s get backup running process id failed: %v", c.FaultType, err)
	}

	pid, err := strconv.Atoi(strings.Trim(pidStr, "\n"))
	if err != nil {
		return fmt.Errorf("%s trans pid string to int failed: %v", c.FaultType, err)
	}
	if err = syscall.Kill(pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("%s kill backup running process failed: %v", c.FaultType, err)
	}

	// 后台发送信号进程退出时可能已经向目标进程发送SIGSTOP，确保被故障注入的程序能够正常运行，
	// 重新发送一次SIGCONT信号。
	if err := syscall.Kill(c.pid, syscall.SIGCONT); err != nil {
		return fmt.Errorf("send signal: SIGCONT to %d failed: %v", c.pid, err)
	}
	return nil
}
