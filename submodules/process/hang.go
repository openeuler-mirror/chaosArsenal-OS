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
	var newFaultType = hang{
		FaultType: "process-hang",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type hang struct {
	FaultType string
	flags     map[string]string
	pid       int
}

func (h *hang) Prepare(inputArgs []string) error {
	h.flags = parse.TransInputFlagsToMap(inputArgs)
	pid, err := GetProcessPidAndExistCheck(h.flags)
	if err != nil {
		return err
	}
	h.pid = pid
	return nil
}

func (h *hang) FaultInject(_ []string) error {
	if err := syscall.Kill(h.pid, syscall.SIGSTOP); err != nil {
		return fmt.Errorf("stop process: %d failed: %v", h.pid, err)
	}
	return nil
}

func (h *hang) FaultRemove(_ []string) error {
	if err := syscall.Kill(h.pid, syscall.SIGCONT); err != nil {
		return fmt.Errorf("run process %d failed: %v", h.pid, err)
	}
	return nil
}
