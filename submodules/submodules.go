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

package submodules

import (
	"fmt"
)

type FaultOperationType func(faultType FaultOperations, inputArgs []string) error

var (
	// OpsTypeIndex 操作类型在输入参数中的索引。
	OpsTypeIndex = 1
	// ModuleNameIndex 模块名在输入参数中的索引。
	ModuleNameIndex = 2
	// FaultTypeIndex 故障模式在输入参数中的索引。
	FaultTypeIndex = 3
	// Inject 故障注入字符串标志。
	Inject = "inject"
	// Remove 故障清理字符串标志。
	Remove = "remove"
	// FaultOperationTypes 故障操作类型集合。
	FaultOperationTypes = map[string]FaultOperationType{}
	// FaultTypes 故障模式对应处理函数集合。
	FaultTypes = map[string]FaultOperations{}
)

// Add 向故障模式处理函数集合中添加元素。
func Add(name string, newFaultType FaultOperations) {
	FaultTypes[name] = newFaultType
}

type FaultOperations interface {
	// Prepare 操作前的准备工作。
	Prepare([]string) error
	// FaultInject 故障注入入口。
	FaultInject([]string) error
	// FaultRemove 故障清除入口。
	FaultRemove([]string) error
}

func RunCmd(inputArgs []string) error {
	// 检查是否支持对应的faultType。
	faultTypeKey := fmt.Sprintf("%s-%s", inputArgs[ModuleNameIndex], inputArgs[FaultTypeIndex])
	handler, ok := FaultTypes[faultTypeKey]
	if !ok {
		return fmt.Errorf("unsupported fault type: %s", faultTypeKey)
	}

	// 如果是阻塞执行先非阻塞执行只执行prepare，做一些前置检查，前置检查不通过肯定是失败的。
	if err := handler.Prepare(inputArgs); err != nil {
		return err
	}

	if inputArgs[OpsTypeIndex] == "prepare" {
		return nil
	}
	ops, ok := FaultOperationTypes[inputArgs[OpsTypeIndex]]
	if !ok {
		return fmt.Errorf("unsupported operation type: %s", inputArgs[ModuleNameIndex])
	}
	return ops(handler, inputArgs)
}
