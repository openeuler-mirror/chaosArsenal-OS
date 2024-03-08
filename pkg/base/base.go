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

package base

import (
	"fmt"

	"arsenal-os/submodules"
	// 初始化opsType和故障注入接口map。
	_ "arsenal-os/submodules/all"
)

// Run 运行故障注入原子能力。
func Run(args []string) error {
	var minimumInputArgs = 4
	// 在cobra中已经做了参数校验，只做简单参数个数校验。
	if len(args) < minimumInputArgs {
		return fmt.Errorf("invalid input parameter")
	}
	return submodules.RunCmd(args)
}
