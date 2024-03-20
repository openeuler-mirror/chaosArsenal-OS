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

	"arsenal-os/util"
)

// processIsExist 检查进程是否存在。
func processIsExist(pid int) bool {
	return util.FileIsExist(fmt.Sprintf("/proc/%d", pid))
}

// GetProcessPidAndExistCheck 检查输入参数pid对应进程是否存在。
func GetProcessPidAndExistCheck(flagsMap map[string]string) (int, error) {
	if _, ok := flagsMap["pid"]; !ok {
		return -1, errors.New("please input params: pid")
	}
	pid, err := strconv.Atoi(flagsMap["pid"])
	if err != nil {
		return -1, fmt.Errorf("trans pid string to int failed: %v", err)
	}
	if !processIsExist(pid) {
		return pid, fmt.Errorf("the process: %d does not exist", pid)
	}
	return pid, nil
}
