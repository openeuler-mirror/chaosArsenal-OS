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

package util

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// FileIsExist 判断文件是否存在。
func FileIsExist(path string) bool {
	_, ret := os.Stat(path)
	return ret == nil
}

// ExecCommandBlock 阻塞执行shell命令，默认的超时时间为5秒。
func ExecCommandBlock(shellCmd string, timeout ...uint64) (string, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	if len(timeout) > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), (time.Duration(timeout[0]))*(time.Second))
	} else {
		const defaultTimeout = 5
		ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout*time.Second)
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", shellCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("execute command: %s timeout, default 5s", shellCmd)
	}

	return out.String(), err
}

// ExecCommandUnblock 非阻塞执行shell命令。
func ExecCommandUnblock(shellCmd string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", shellCmd)
	if err := cmd.Start(); err != nil {
		return "", err
	}
	return strconv.Itoa(cmd.Process.Pid), nil
}

// CheckEnvShellCommand shell命令依赖检查。
func CheckEnvShellCommand(commands []string) ([]string, bool) {
	missingCommands := make([]string, 0)
	for _, value := range commands {
		if _, err := ExecCommandBlock(fmt.Sprintf("type %s", value)); err != nil {
			missingCommands = append(missingCommands, value)
		}
	}

	if len(missingCommands) != 0 {
		return missingCommands, true
	}
	return missingCommands, false
}
