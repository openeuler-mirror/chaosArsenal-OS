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

package util

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	// from /usr/include/linux/fs.h
	FS_SECRM_FL        = 0x00000001 /* Secure deletion */
	FS_UNRM_FL         = 0x00000002 /* Undelete */
	FS_COMPR_FL        = 0x00000004 /* Compress file */
	FS_SYNC_FL         = 0x00000008 /* Synchronous updates */
	FS_IMMUTABLE_FL    = 0x00000010 /* Immutable file */
	FS_APPEND_FL       = 0x00000020 /* writes to file may only append */
	FS_NODUMP_FL       = 0x00000040 /* do not dump file */
	FS_NOATIME_FL      = 0x00000080 /* do not update atime */
	FS_DIRTY_FL        = 0x00000100
	FS_COMPRBLK_FL     = 0x00000200 /* One or more compressed clusters */
	FS_NOCOMP_FL       = 0x00000400 /* Don't compress */
	FS_ENCRYPT_FL      = 0x00000800 /* Encrypted file */
	FS_BTREE_FL        = 0x00001000 /* btree format dir */
	FS_INDEX_FL        = 0x00001000 /* hash-indexed directory */
	FS_IMAGIC_FL       = 0x00002000 /* AFS directory */
	FS_JOURNAL_DATA_FL = 0x00004000 /* Reserved for ext3 */
	FS_NOTAIL_FL       = 0x00008000 /* file tail should not be merged */
	FS_DIRSYNC_FL      = 0x00010000 /* dirsync behaviour (directories only) */
	FS_TOPDIR_FL       = 0x00020000 /* Top of directory hierarchies*/
	FS_HUGE_FILE_FL    = 0x00040000 /* Reserved for ext4 */
	FS_EXTENT_FL       = 0x00080000 /* Extents */
	FS_EA_INODE_FL     = 0x00200000 /* Inode used for large EA */
	FS_EOFBLOCKS_FL    = 0x00400000 /* Reserved for ext4 */
	FS_NOCOW_FL        = 0x00800000 /* Do not cow file */
	FS_INLINE_DATA_FL  = 0x10000000 /* Reserved for ext4 */
	FS_PROJINHERIT_FL  = 0x20000000 /* Create with parents projid */
	FS_RESERVED_FL     = 0x80000000 /* reserved for ext2 lib */
)

// Request flags.
const (
	// from ioctl_list manpage
	FS_IOC_GETFLAGS uintptr = 0x80086601
	FS_IOC_SETFLAGS uintptr = 0x40086602
)

func ioctl(f *os.File, request uintptr, attrp *int32) error {
	argp := uintptr(unsafe.Pointer(attrp))
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), request, argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}

	return nil
}

// GetAttrs retrieves the attributes of a file.
func GetAttrs(f *os.File) (int32, error) {
	attr := int32(-1)
	err := ioctl(f, FS_IOC_GETFLAGS, &attr)
	return attr, err
}

// SetAttr sets the given attribute.
func SetAttr(f *os.File, attr int32) error {
	attrs, err := GetAttrs(f)
	if err != nil {
		return err
	}

	attrs |= attr
	return ioctl(f, FS_IOC_SETFLAGS, &attrs)
}

// UnsetAttr unsets the given attribute.
func UnsetAttr(f *os.File, attr int32) error {
	attrs, err := GetAttrs(f)
	if err != nil {
		return err
	}

	attrs ^= (attrs & attr)
	return ioctl(f, FS_IOC_SETFLAGS, &attrs)
}

// IsAttr checks whether the given attribute is set.
func IsAttr(f *os.File, attr int32) (bool, error) {
	attrs, err := GetAttrs(f)
	if err != nil {
		return false, err
	}

	if (attrs & attr) != 0 {
		return true, nil
	}
	return false, nil
}
