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
	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
)

func init() {
	var newFaultType = serviceStop{
		FaultType: "system-service-stop",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type serviceStop struct {
	FaultType string
	ops       serviceOps
}

func (r *serviceStop) Prepare(inputArgs []string) error {
	flags := parse.TransInputFlagsToMap(inputArgs)
	return r.ops.PreRun(flags, inputArgs[submodules.OpsTypeIndex])
}

func (r *serviceStop) FaultInject(_ []string) error {
	return r.ops.executor("stop")
}

func (r *serviceStop) FaultRemove(_ []string) error {
	return r.ops.executor("start")
}
