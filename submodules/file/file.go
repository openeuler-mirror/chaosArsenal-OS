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
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"syscall"

	"arsenal-os/util"
)

const openFilePerm = os.FileMode(0644)

// getFileFullPath 获取文件的全路径。
func getFileFullPath(filePath string) (string, error) {
	if filepath.IsAbs(filePath) {
		return filePath, nil
	}

	fullPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func getFilePathAndCheck(flags map[string]string) (string, error) {
	if _, ok := flags["path"]; !ok {
		return "", errors.New("please input param: path")
	}
	filePath, err := getFileFullPath(flags["path"])
	if err != nil {
		return "", fmt.Errorf("get file %s full path failed, error: %s", filePath, err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return filePath, fmt.Errorf("please check file %s exist", filePath)
	}
	// 文件类故障只针对于文件。
	if fileInfo.IsDir() {
		return filePath, fmt.Errorf("%s is directory", filePath)
	}
	return filePath, nil
}

func fileAttrOps(filePath string, attr int32, opsType string) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY, openFilePerm)
	if err != nil {
		return fmt.Errorf("open file: %s failed, Err: %s", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("close file %s failed, Error: %s", filePath, err)
		}
	}()

	switch opsType {
	case "set":
		err = util.SetAttr(file, attr)
	case "unset":
		err = util.UnsetAttr(file, attr)
	default:
		return fmt.Errorf("unsupported file attr set type")
	}
	if err != nil {
		return fmt.Errorf("set file: %s attr as %d failed", filePath, attr)
	}
	return nil
}

// getFileSize 获取给定路径文件size，单位为字节。
func getFileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// getRandomString 返回一个指定长度的随机字符串。
func getRandomString(length int64) (string, error) {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength := big.NewInt(int64(len(charset)))
	randomString := make([]byte, length)
	for index := int64(0); index < length; index++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		randomString[index] = charset[randomIndex.Int64()]
	}

	return string(randomString), nil
}

// getPathMountPointAvailSize 获取输入文件路径对应挂载点剩余可用磁盘空间。
func getPathMountPointAvailSize(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("get path avail size failed: %v", err)
	}
	return stat.Bsize * int64(stat.Bavail), nil
}

// fileCopy 文件拷贝。
func fileCopy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	if _, err = os.Stat(dst); err == nil {
		return 0, fmt.Errorf("%s file is exist", dst)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
