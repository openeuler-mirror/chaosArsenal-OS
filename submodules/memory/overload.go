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

package memory

import (
	"fmt"

	"arsenal-os/pkg/tools"
	"arsenal-os/submodules"
)

func init() {
	var newFaultType = overload{
		FaultType: "memory-overload",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type overload struct {
	FaultType string
	stressNg  tools.StressNg
}

func (o *overload) Prepare(inputArgs []string) error {
	// stress-ng添加持续消耗系统内存参数私有参数，
	// --vm-keep 不做map和unmap操作，申请内存不释放，持续写内存。
	// --vm-populate 先消耗普通内存，当普通内存不足时，消耗swap内存。
	if err := o.stressNg.PreRun(inputArgs, "--vm-keep --vm-populate"); err != nil {
		return fmt.Errorf("run stress-ng test program failed: %v", err)
	}
	return nil
}

func (o *overload) FaultInject(_ []string) error {
	if err := o.stressNg.Run(); err != nil {
		return fmt.Errorf("inject %s failed: %v", o.FaultType, err)
	}
	return nil
}

func (o *overload) FaultRemove(_ []string) error {
	if err := o.stressNg.Destroy(); err != nil {
		return fmt.Errorf("remove %s failed: %v", o.FaultType, err)
	}
	return nil
}
