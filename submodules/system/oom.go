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
	var newFaultType = oom{
		FaultType: "system-oom",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type oom struct {
	FaultType string
}

func (o *oom) Prepare(_ []string) error {
	return triggerRunEnvChecker()
}

func (o *oom) FaultInject(_ []string) error {
	if result, err := util.ExecCommandBlock(fmt.Sprintf("echo f > %s", Trigger)); err != nil {
		return fmt.Errorf("make system %s failed, err: %v, result: %s", o.FaultType, err, result)
	}
	return nil
}

func (o *oom) FaultRemove(_ []string) error {
	return nil
}
