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

package cpu

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = offline{
		FaultType: "cpu-offline",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type offline struct {
	FaultType string
	flags     map[string]string
	cpuList   []int
}

func (o *offline) isDuplicateCPUID(checkID int64) bool {
	for _, id := range o.cpuList {
		if id == int(checkID) {
			return true
		}
	}
	return false
}

func (o *offline) cpuListParser() error {
	cpuListString, ok := o.flags["cpuid"]
	if !ok {
		return errors.New("please input param cpuid")
	}
	re := regexp.MustCompile(`^(\d+(-\d+)?)(,\d+(-\d+)?)*$`)
	if !re.MatchString(cpuListString) {
		return fmt.Errorf("cpuid param format error: %s", cpuListString)
	}

	// 将字符串按逗号分隔成多个子串。
	parts := strings.Split(cpuListString, ",")
	for _, part := range parts {
		// 判断子串中是否包含连字符。
		if strings.Contains(part, "-") {
			// 将子串按连字符分隔成两个数字。
			rangeParts := strings.Split(part, "-")
			start, err := strconv.ParseInt(rangeParts[0], 10, 64)
			if err != nil {
				return fmt.Errorf("trans cpu range starting id to int failed: %v", err)
			}
			end, err := strconv.ParseInt(rangeParts[1], 10, 64)
			if err != nil {
				return fmt.Errorf("trans cpu range ending id to int failed: %v", err)
			}
			if start > end {
				return fmt.Errorf("cpu range starting id is larger than ending id: %s", part)
			}
			// 将两个数字之间的所有整数加入数组。
			for i := start; i <= end; i++ {
				if o.isDuplicateCPUID(i) {
					return fmt.Errorf("duplicate cpu id: %d", i)
				}
				o.cpuList = append(o.cpuList, int(i))
			}
		} else {
			// 将子串转换为整数并加入数组。
			num, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return fmt.Errorf("trans single cpu id string to int failed: %v", err)
			}
			if o.isDuplicateCPUID(num) {
				return fmt.Errorf("duplicate cpu id: %d", num)
			}
			o.cpuList = append(o.cpuList, int(num))
		}
	}
	return nil
}

func (o *offline) cpuExistenceCheck() error {
	for _, cpuID := range o.cpuList {
		offlineCtlPath := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/online", cpuID)
		if !util.FileIsExist(offlineCtlPath) {
			return fmt.Errorf("can not found offline control file path: %s", offlineCtlPath)
		}
	}
	return nil
}

func (o *offline) Prepare(inputArgs []string) error {
	if missingCmd, isMissCmd := util.CheckEnvShellCommand([]string{"echo"}); isMissCmd {
		return fmt.Errorf("missing command: %s", missingCmd)
	}

	o.flags = parse.TransInputFlagsToMap(inputArgs)
	if err := o.cpuListParser(); err != nil {
		return fmt.Errorf("parser cpu id failed: %v", err)
	}

	if err := o.cpuExistenceCheck(); err != nil {
		return fmt.Errorf("cpu existence check failed: %v", err)
	}
	return nil
}

func (o *offline) executor(magic string) error {
	for _, cpuID := range o.cpuList {
		offlineCtlPath := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/online", cpuID)
		shellCmd := fmt.Sprintf("echo %s > %s", magic, offlineCtlPath)
		if result, err := util.ExecCommandBlock(shellCmd); err != nil {
			return fmt.Errorf("execute %s failed, error: %s, result: %s", shellCmd, err, result)
		}
	}
	return nil
}

func (o *offline) FaultInject(_ []string) error {
	return o.executor("0")
}

func (o *offline) FaultRemove(_ []string) error {
	return o.executor("1")
}
