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

package file

import (
	"fmt"
	"os"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = readonly{
		FaultType: "file-readonly",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type readonly struct {
	FaultType string
	flags     map[string]string
	filePath  string
	fileAttr  int32
}

func (r *readonly) storeFileRawAttr(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY, openFilePerm)
	if err != nil {
		return fmt.Errorf("open file: %s failed, Err: %s", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("close file %s failed, Error: %s", filePath, err)
		}
	}()

	// 获取文件属性，用于判断文件是否已经处于只读状态。
	fileAttr, err := util.GetAttrs(file)
	if err != nil {
		return fmt.Errorf("get file %s attribute failed, Err: %s", filePath, err)
	}
	r.fileAttr = fileAttr
	return nil
}

func (r *readonly) Prepare(inputArgs []string) error {
	r.flags = parse.TransInputFlagsToMap(inputArgs)
	filePath, err := getFilePathAndCheck(r.flags)
	if err != nil {
		return fmt.Errorf("%s get full path or check error", r.FaultType)
	}
	r.filePath = filePath

	if err = r.storeFileRawAttr(filePath); err != nil {
		return fmt.Errorf("get file: %s attr failed", r.filePath)
	}
	return nil
}

func (r *readonly) FaultInject(_ []string) error {
	immutableFlag := r.fileAttr & util.FS_IMMUTABLE_FL
	if immutableFlag == util.FS_IMMUTABLE_FL {
		return fmt.Errorf("the file %s is already in an not-writable state", r.filePath)
	}

	if err := fileAttrOps(r.filePath, util.FS_IMMUTABLE_FL, "set"); err != nil {
		return fmt.Errorf("set file attr failed: %v", err)
	}
	return nil
}

func (r *readonly) FaultRemove(_ []string) error {
	if err := fileAttrOps(r.filePath, util.FS_IMMUTABLE_FL, "unset"); err != nil {
		return fmt.Errorf("unset file attr failed: %v", err)
	}
	return nil
}
