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

package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

var stressNgPath = "third_party_tools/stress-ng"

// StressNg 用于记录工具stress-ng命令相关信息。
type StressNg struct {
	FullPath       string
	StressNgCmd    string
	nice           string
	pidSearchNgCmd string
}

// stressNgExePermCheck 检查stress-ng命令是否有可执行权限，如果没有则赋予可执行权限。
func (s *StressNg) stressNgExecPermCheck() error {
	arsenalOsPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get arsenal-os absolute path failed(%v)", err)
	}

	stressNgFullPath := fmt.Sprintf("%s/%s", filepath.Dir(arsenalOsPath), stressNgPath)
	fileInfo, err := os.Stat(stressNgFullPath)
	if err != nil {
		return fmt.Errorf("get %s info failed: %v", stressNgPath, err)
	}
	const fileExecPermission = 0755
	if fileInfo.Mode()&0111 == 0 {
		if err := os.Chmod(stressNgFullPath, fileExecPermission); err != nil {
			return fmt.Errorf("chmod %s failed: %v", stressNgFullPath, err)
		}
	}
	s.FullPath = stressNgFullPath
	return nil
}

func (s *StressNg) setRunNice(inputArgs []string) {
	flags := parse.TransInputFlagsToMap(inputArgs)
	if nice, ok := flags["nice"]; ok {
		s.nice = nice
	}
}

// setRunCliCmd privateArgs用于传入stress-ng特定参数，如--vm-keep --vm-populate。
func (s *StressNg) setRunCliCmd(inputArgs []string, privateArgs ...string) {
	// 将传入privateArgs数组类型转换成字符串类型。
	privateArgsStr := strings.Join(privateArgs, "")
	flagsString := parse.TransInputFlagsToString(inputArgs)

	// 如果输入的args中含有nice字段，需要将nice字段从args中移除，nice非stress-ng参数，
	// 该场景通过nice命令设定stress-ng进程优先级。
	re := regexp.MustCompile(`--nice\s+-?\d+`)
	flagsString = re.ReplaceAllString(flagsString, "")

	// 1、--nice在flagsString字符串中间替换后将出现双空格；
	// 2、--nice在flagsString字符串末尾替换后末尾将多出空格；
	// 以上两种情况都会影响故障清理时找不到对应pid，因为pid查找通过命令全词匹配，需要做处理。
	flagsString = strings.ReplaceAll(flagsString, "  ", " ")
	flagsString = strings.TrimRight(flagsString, " ")

	if s.nice != "" {
		var pidSearchNgCmd string
		if len(privateArgs) != 0 {
			s.StressNgCmd = fmt.Sprintf("nice -n %s %s %s %s",
				s.nice, s.FullPath, flagsString, privateArgs)
			pidSearchNgCmd = fmt.Sprintf("%s %s %s", s.FullPath, flagsString, privateArgs)
		} else {
			s.StressNgCmd = fmt.Sprintf("nice -n %s %s %s",
				s.nice, s.FullPath, flagsString)
			pidSearchNgCmd = fmt.Sprintf("%s %s", s.FullPath, flagsString)
		}
		s.pidSearchNgCmd = pidSearchNgCmd
	} else {
		var shellCmd string
		if len(privateArgs) != 0 {
			shellCmd = fmt.Sprintf("%s %s %s", s.FullPath, flagsString, privateArgsStr)
		} else {
			shellCmd = fmt.Sprintf("%s %s", s.FullPath, flagsString)
		}
		s.StressNgCmd = shellCmd
	}
}

// PreRun 依赖检查，预运行stress-ng命令，验证stress-ng命令的正确性。
func (s *StressNg) PreRun(inputArgs []string, privateArgs ...string) error {
	dependCmd := []string{"kill", "ps", "grep", "awk"}
	if missingCmd, isMissCmd := util.CheckEnvShellCommand(dependCmd); isMissCmd {
		return fmt.Errorf("missing command: %s", missingCmd)
	}

	if err := s.stressNgExecPermCheck(); err != nil {
		return err
	}
	s.setRunNice(inputArgs)
	s.setRunCliCmd(inputArgs, privateArgs...)

	// 故障清理场景不需要执行预运行，直接返回nil。
	if inputArgs[submodules.OpsTypeIndex] == submodules.Remove {
		return nil
	}
	// 先设定4s的运行时间，根据返回信息判断命令是否可以正常运行。
	stressNgTestCmd := fmt.Sprintf("%s -t 4s", s.StressNgCmd)
	if result, err := util.ExecCommandBlock(stressNgTestCmd); err != nil {
		return fmt.Errorf("execute command: %s failed, err: %v result: %s",
			stressNgTestCmd, err, result)
	}
	return nil
}

// Run 运行stress-ng命令。
func (s *StressNg) Run() error {
	if result, err := util.ExecCommandUnblock(s.StressNgCmd); err != nil {
		return fmt.Errorf("execute command: %s failed, err: %v result: %s", s.StressNgCmd, err, result)
	}
	return nil
}

// Destroy 全词匹配的方式查找后台运行stress-ng相关进程pid后，将对应进程kill掉。
func (s *StressNg) Destroy() error {
	var searchStr string
	if s.nice != "" {
		searchStr = s.pidSearchNgCmd
	} else {
		searchStr = s.StressNgCmd
	}
	getPidShellCmd := fmt.Sprintf("ps aux | grep -v grep | grep '%s' | awk '{print $2}'", searchStr)
	pidStr, err := util.ExecCommandBlock(getPidShellCmd)
	if err != nil || pidStr == "" {
		return fmt.Errorf("failed to obtain pid of stress-ng process running in the background")
	}

	// 可能存在多个stress-ng进程，需要kill掉所有进程。
	killCmd := fmt.Sprintf("kill -9 %s", strings.ReplaceAll(pidStr, "\n", " "))
	if result, err := util.ExecCommandBlock(killCmd); err != nil {
		return fmt.Errorf("execute command: %s failed, err: %v result: %s", killCmd, err, result)
	}
	return nil
}
