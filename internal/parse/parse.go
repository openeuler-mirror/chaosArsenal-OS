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

package parse

import (
	"fmt"
	"sort"
	"strings"
)

// minimumFlagLength 输入参数以--开头，最短字符串长度为3。
const minimumFlagLength = 3

func orderFlagsMapKey(inputMap map[string]string) []string {
	// 将map的key转换成切片。
	keys := make([]string, 0, len(inputMap))
	for key := range inputMap {
		keys = append(keys, key)
	}

	// 对切片进行排序。
	sort.Strings(keys)
	return keys
}

// TransInputFlagsToMap 将输入参数转成map，如：map[cpu:1 sched-prio:2]。
func TransInputFlagsToMap(args []string) map[string]string {
	flags := make(map[string]string)
	for index, flag := range args {
		// Example flag: --path
		if len(flag) < minimumFlagLength || index+1 >= len(args) {
			continue
		}

		if strings.HasPrefix(flag, "--") {
			flagName := strings.TrimPrefix(flag, "--")
			flags[flagName] = args[index+1]
		}
	}
	return flags
}

// TransInputFlagsToString 将输入的参数转换成有序的flags字符串，如：--cpu 1 --sched-prio 2。
func TransInputFlagsToString(args []string) string {
	var flagsString string
	flags := TransInputFlagsToMap(args)

	// 获取到参数map是无序的，有些故障清理时依赖flags的顺序，所以需要对map进行排序。
	for _, key := range orderFlagsMapKey(flags) {
		flagsString += fmt.Sprintf("--%s %s ", key, flags[key])
	}

	if len(flagsString) > minimumFlagLength {
		flagsString = strings.TrimSpace(flagsString)
	}
	return flagsString
}
