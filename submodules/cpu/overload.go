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
	"fmt"

	"arsenal-os/pkg/tools"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = overload{
		FaultType: "cpu-overload",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type overload struct {
	FaultType string
	stressNg  tools.StressNg
}

func (o *overload) Prepare(inputArgs []string) error {
	if missingCmd, isMissCmd := util.CheckEnvShellCommand([]string{"nice"}); isMissCmd {
		return fmt.Errorf("missing command: %s", missingCmd)
	}

	if err := o.stressNg.PreRun(inputArgs); err != nil {
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
