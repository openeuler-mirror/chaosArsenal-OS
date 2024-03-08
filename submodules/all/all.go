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

package all

import (
	// 添加注入清理接口
	_ "arsenal-os/internal/operations"
	// 向全局故障相关操作接口map中添加cpu类型接口
	_ "arsenal-os/submodules/cpu"
	// 向全局故障相关操作接口map中添加file类型接口
	_ "arsenal-os/submodules/file"
	// 向全局故障相关操作接口map中添加filesystem类型接口
	_ "arsenal-os/submodules/filesystem"
	// 向全局故障相关操作接口map中添加memory类型接口
	_ "arsenal-os/submodules/memory"
	// 向全局故障相关操作接口map中添加process类型接口
	_ "arsenal-os/submodules/process"
	// 向全局故障相关操作接口map中添加system类型接口
	_ "arsenal-os/submodules/system"
)
