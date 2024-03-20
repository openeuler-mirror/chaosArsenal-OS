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
	"fmt"
	"syscall"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
)

func init() {
	var newFaultType = exitAbnormally{
		FaultType: "process-exit-abnormal",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type exitAbnormally struct {
	FaultType string
	flags     map[string]string
	pid       int
}

func (e *exitAbnormally) Prepare(inputArgs []string) error {
	e.flags = parse.TransInputFlagsToMap(inputArgs)
	pid, err := GetProcessPidAndExistCheck(e.flags)
	if err != nil {
		return err
	}
	e.pid = pid
	return nil
}

func (e *exitAbnormally) FaultInject(_ []string) error {
	if err := syscall.Kill(e.pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("%s kill process: %d failed: %v", e.FaultType, e.pid, err)
	}
	return nil
}

func (e *exitAbnormally) FaultRemove(_ []string) error {
	return nil
}
