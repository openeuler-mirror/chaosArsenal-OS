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

	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = rebootAbnormal{
		FaultType: "system-reboot-abnormal",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type rebootAbnormal struct {
	FaultType string
}

func (r *rebootAbnormal) Prepare(_ []string) error {
	return triggerRunEnvChecker()
}

func (r *rebootAbnormal) FaultInject(_ []string) error {
	if result, err := util.ExecCommandBlock(fmt.Sprintf("echo b > %s", Trigger)); err != nil {
		return fmt.Errorf("make system %s failed, err: %v, result: %s", r.FaultType, err, result)
	}
	return nil
}

func (r *rebootAbnormal) FaultRemove(_ []string) error {
	return nil
}
