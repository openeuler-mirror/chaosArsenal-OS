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

package system

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"arsenal-os/util"
)

var Trigger = "/proc/sysrq-trigger"

func triggerRunEnvChecker() error {
	if missingCmd, isMissCmd := util.CheckEnvShellCommand([]string{"echo"}); isMissCmd {
		return fmt.Errorf("missing command: %s", missingCmd)
	}

	if !util.FileIsExist(Trigger) {
		return fmt.Errorf("can't found file: %s", Trigger)
	}
	return nil
}

// parseDuration 解析输入时间字符串，返回一个time.Duration类型的值，
// 字符串格式为：1h:1m:1s，时分秒以':'隔开。
func parseDuration(input string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+h)?:?(\d+m)?:?(\d+s)?$`)
	if !re.MatchString(input) {
		return 0, fmt.Errorf("invalid input: %s", input)
	}

	var duration time.Duration
	parts := strings.Split(input, ":")
	for _, part := range parts {
		// 获取倒数第一个字符为单位。
		unit := part[len(part)-1]
		value, err := strconv.Atoi(part[:len(part)-1])
		if err != nil {
			return 0, err
		}
		switch unit {
		case 'h':
			duration += time.Duration(value) * time.Hour
		case 'm':
			duration += time.Duration(value) * time.Minute
		case 's':
			duration += time.Duration(value) * time.Second
		default:
			return 0, fmt.Errorf("invalid trans time unit: %c", unit)
		}
	}
	return duration, nil
}
