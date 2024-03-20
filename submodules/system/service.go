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
	"errors"
	"fmt"

	"arsenal-os/submodules"
	"arsenal-os/util"
)

type serviceOps struct {
	serviceName string
}

func (s *serviceOps) getOpsCmd(opsType string) string {
	return fmt.Sprintf("service %s %s", s.serviceName, opsType)
}

func (s *serviceOps) checkServiceStatus() error {
	checkCmd := s.getOpsCmd("status")
	if _, err := util.ExecCommandBlock(checkCmd); err != nil {
		switch err.Error() {
		case "exit status 3":
			return fmt.Errorf("the service %s is in inactive status", s.serviceName)
		case "exit status 4":
			return fmt.Errorf("no such service %s", s.serviceName)
		default:
			return fmt.Errorf("execute command: %s failed, err: %v", checkCmd, err)
		}
	}
	return nil
}

func (s *serviceOps) PreRun(flags map[string]string, opsType string) error {
	serviceName, ok := flags["name"]
	if !ok {
		return errors.New("service name is required")
	}
	s.serviceName = serviceName
	// 只有在注入操作的场景下需要考虑服务是不是已经处于stop状态。
	if opsType == submodules.Inject {
		if err := s.checkServiceStatus(); err != nil {
			return fmt.Errorf("check service status failed(%s)", err)
		}
	}
	return nil
}

func (s *serviceOps) executor(opsType string) error {
	shellCmd := s.getOpsCmd(opsType)
	if result, err := util.ExecCommandBlock(shellCmd); err != nil {
		return fmt.Errorf("execute command: %s failed, err: %v result: %s", shellCmd, err, result)
	}
	return nil
}
