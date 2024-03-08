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

package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"arsenal-os/internal/parse"
	"arsenal-os/submodules"
	"arsenal-os/util"
)

func init() {
	var newFaultType = mountPointInodeExhaustion{
		FaultType: "filesystem-mountpoint-inode-exhaustion",
	}
	submodules.Add(newFaultType.FaultType, &newFaultType)
}

type mountPointInodeExhaustion struct {
	FaultType   string
	flags       map[string]string
	mountPoint  string
	testFileDir string
	exitChan    chan int
}

var (
	numOfDir int
	lock     sync.Mutex
	dirList  []string
	// numberOfCoroutines 文件创建和删除的最大协程数量。
	numberOfCoroutines = runtime.NumCPU() * 2
)

func (m *mountPointInodeExhaustion) Prepare(inputArgs []string) error {
	if missingCmd, isMissCmd := util.CheckEnvShellCommand([]string{"df"}); isMissCmd {
		return fmt.Errorf("missing command: %s", missingCmd)
	}

	m.flags = parse.TransInputFlagsToMap(inputArgs)
	mountPoint, err := mountPointCheck(m.flags)
	if err != nil {
		return fmt.Errorf("mount point check failed: %v", err)
	}
	m.mountPoint = mountPoint
	m.exitChan = make(chan int, numberOfCoroutines)
	m.testFileDir = fmt.Sprintf("%s/arsenal_test_dir/", m.mountPoint)
	return nil
}

func (m *mountPointInodeExhaustion) FaultInject(_ []string) error {
	if !util.FileIsExist(m.testFileDir) {
		const createTestFilePerm = os.FileMode(0644)
		if err := os.Mkdir(m.testFileDir, createTestFilePerm); err != nil {
			return fmt.Errorf("make test dir: %s failed", m.testFileDir)
		}
	}

	for {
		for j := 0; j < numberOfCoroutines; j++ {
			go m.createFile()
		}

		// 等待文件创建的所有协程退出。
		for j := 0; j < numberOfCoroutines; j++ {
			<-m.exitChan
		}

		// 判断挂载点inode有没有耗尽，如果耗尽停止创建消耗协程。
		if m.mntInodeIsFree() {
			break
		}
	}
	return nil
}

func (m *mountPointInodeExhaustion) killBackgroundInjectProcess(inputArgs []string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get execute binary file path failed(%v)", err)
	}

	searchStr := fmt.Sprintf("%s inject %s %s --path %s", exePath,
		inputArgs[submodules.ModuleNameIndex], inputArgs[submodules.FaultTypeIndex],
		m.mountPoint)
	getPidShellCmd := fmt.Sprintf("ps aux | grep -v grep | grep -w '%s' | awk '{print $2}'", searchStr)
	pidStr, err := util.ExecCommandBlock(getPidShellCmd)
	if err != nil {
		return fmt.Errorf("failed to obtain pid of arsenal-os %s inject process "+
			"running in the background", m.FaultType)
	}
	if pidStr == "" {
		return nil
	}

	killCmd := fmt.Sprintf("kill -9 %s", strings.ReplaceAll(pidStr, "\n", " "))
	if result, err := util.ExecCommandBlock(killCmd); err != nil {
		return fmt.Errorf("execute command: %s failed, err: %v result: %s", killCmd, err, result)
	}
	return nil
}

// FaultRemove 每个线程删除一个创建的文件夹。
func (m *mountPointInodeExhaustion) FaultRemove(inputArgs []string) error {
	// 存在用户提前做清理的场景，需要将inode创建后台执行进程kill掉。
	if err := m.killBackgroundInjectProcess(inputArgs); err != nil {
		return fmt.Errorf("%s kill background inject process failed(%v)", m.FaultType, err)
	}

	dirNum := m.getTestDirNum()
	cycle := dirNum / numberOfCoroutines
	remain := dirNum % numberOfCoroutines

	for j := 0; j < cycle; j++ {
		for k := 0; k < numberOfCoroutines; k++ {
			dirNum--
			go deleteFIle(dirList[dirNum], m.exitChan)
		}

		// 等待文件删除的所有协程退出。
		for num := 0; num < numberOfCoroutines; num++ {
			<-m.exitChan
		}
	}

	if remain != 0 {
		for j := 0; j < remain; j++ {
			dirNum--
			go deleteFIle(dirList[dirNum], m.exitChan)
		}

		for k := 0; k < remain; k++ {
			<-m.exitChan
		}
	}

	if !util.FileIsExist(m.testFileDir) {
		return nil
	}
	if err := os.RemoveAll(m.testFileDir); err != nil {
		fmt.Printf("remove dir: %s failed: %v\n", m.testFileDir, err)
		os.Exit(-1)
	}
	return nil
}

func (m *mountPointInodeExhaustion) createFile() {
	lock.Lock()
	numOfDir++
	numOfDirStr := strconv.Itoa(numOfDir)
	lock.Unlock()
	fileDir := fmt.Sprintf("%s/arsenal_test_dir/arsenal_dir_%s", m.mountPoint, numOfDirStr)
	const createTestDirPerm = os.FileMode(0666)
	if err := os.Mkdir(fileDir, createTestDirPerm); err != nil {
		if strings.Contains(fmt.Sprintf("%s", err), "no space left on device") {
			m.exitChan <- 1
		}
	}

	// TODO: 动态获取某个文件系统的文件夹下可创建文件的最大数量。
	for j := 0; j < 10000; j++ {
		numOfFileStr := strconv.Itoa(j)
		file, err := os.Create(fmt.Sprintf("%s/test_%s", fileDir, numOfFileStr))
		if err != nil {
			if strings.Contains(fmt.Sprintf("%s", err), "no space left on device") {
				m.exitChan <- 1
				break
			}
		}
		defer file.Close()
	}
	m.exitChan <- 0
}

func (m *mountPointInodeExhaustion) getTestDirNum() int {
	var dirNum int
	testFileDir := fmt.Sprintf("%s/arsenal_test_dir/", m.mountPoint)
	fileList, _ := ioutil.ReadDir(testFileDir)
	for _, value := range fileList {
		dirNum++
		dirList = append(dirList, fmt.Sprintf("%s%s", testFileDir, value.Name()))
	}
	return dirNum
}

func deleteFIle(dirFullPath string, exitChan chan int) {
	fileList, _ := ioutil.ReadDir(dirFullPath)
	for _, value := range fileList {
		if err := os.Remove(fmt.Sprintf("%s/%s", dirFullPath, value.Name())); err != nil {
			panic(err)
		}
	}

	if err := os.RemoveAll(dirFullPath); err != nil {
		fmt.Printf("Remove file: %s failed: %v", dirFullPath, err)
	}
	exitChan <- 1
}

func (m *mountPointInodeExhaustion) mntInodeIsFree() bool {
	var statFs syscall.Statfs_t
	if err := syscall.Statfs(m.mountPoint, &statFs); err != nil {
		fmt.Printf("Cannot stat %s: %v\n", m.mountPoint, err)
		os.Exit(1)
	}

	if statFs.Ffree == 0 {
		return true
	}
	return false
}
