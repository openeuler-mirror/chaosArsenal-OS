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
	"time"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = timeJump{
		FaultType: "system-time-jump",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

var (
	validDirections = []string{"backwards", "forwards"}
)

type timeJump struct {
	FaultType string
	direction string
	interval  string
}

func (t *timeJump) Prepare(inputArgs []string) error {
	dependCmd := []string{"date", "hwclock"}
	if missingCmd, isMissCmd := util.CheckEnvShellCommand(dependCmd); isMissCmd {
		return fmt.Errorf("missing command: %s", missingCmd)
	}

	flags := parse.TransInputFlagsToMap(inputArgs)
	direction, ok := flags["direction"]
	if !ok {
		return fmt.Errorf("missing direction flag, example: %s", validDirections)
	}

	isValid := false
	for _, dir := range validDirections {
		if dir == direction {
			isValid = true
			break
		}
	}
	t.direction = direction
	if !isValid {
		return fmt.Errorf("invalid directions %s", direction)
	}

	interval, ok := flags["interval"]
	if !ok {
		return fmt.Errorf("missing interval flag, example: 10s")
	}
	t.interval = interval
	return nil
}

func (t *timeJump) FaultInject(_ []string) error {
	duration, err := parseDuration(t.interval)
	if err != nil {
		return fmt.Errorf("parse duration failed: %v", err)
	}
	now := time.Now()
	if t.direction == "backwards" {
		duration = -duration
	}

	// 当前最大时间跳变粒度为小时。
	newTime := now.Add(duration).Format("15:04:05")
	if result, err := util.ExecCommandBlock(fmt.Sprintf("date -s %s", newTime)); err != nil {
		return fmt.Errorf("make system %s failed: %v, result: %s", t.FaultType, err, result)
	}
	return nil
}

func (t *timeJump) FaultRemove(_ []string) error {
	if result, err := util.ExecCommandBlock("hwclock -s"); err != nil {
		return fmt.Errorf("make system %s failed, err: %v, result: %s", t.FaultType, err, result)
	}
	return nil
}
