//go:build windows

package linker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

type LinkType uint32

const (
	FileSymbolicLink LinkType = iota
	DirectorySymbolicLink
	HardLink
	Junction
)

func (t LinkType) String() string {
	switch t {
	case FileSymbolicLink:
		return "Symbolic link (file)"
	case DirectorySymbolicLink:
		return "Symbolic link (directory)"
	case HardLink:
		return "Hard link"
	case Junction:
		return "Junction"
	default:
		return "Unknown"
	}
}

func Validate(target, linkPath string, typ LinkType) error {
	target = strings.TrimSpace(target)
	linkPath = strings.TrimSpace(linkPath)
	if target == "" {
		return errors.New("target path is required")
	}
	if linkPath == "" {
		return errors.New("link path is required")
	}

	targetAbs, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return fmt.Errorf("resolve target: %w", err)
	}
	linkAbs, err := filepath.Abs(filepath.Clean(linkPath))
	if err != nil {
		return fmt.Errorf("resolve link: %w", err)
	}

	info, err := os.Stat(targetAbs)
	if err != nil {
		return fmt.Errorf("target is not accessible: %w", err)
	}
	if _, err := os.Lstat(linkAbs); err == nil {
		return errors.New("a file or directory already exists at the link path")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check link path: %w", err)
	}
	parentInfo, err := os.Stat(filepath.Dir(linkAbs))
	if err != nil || !parentInfo.IsDir() {
		return errors.New("the destination folder does not exist")
	}
	if err := validateLinkName(filepath.Base(linkAbs)); err != nil {
		return err
	}

	switch typ {
	case FileSymbolicLink:
		if info.IsDir() {
			return errors.New("a file symbolic link requires a file target")
		}
	case DirectorySymbolicLink, Junction:
		if !info.IsDir() {
			return errors.New("this link type requires a directory target")
		}
		if typ == Junction && !sameVolume(targetAbs, linkAbs) {
			return errors.New("a junction target and link must be on the same volume")
		}
	case HardLink:
		if info.IsDir() {
			return errors.New("a hard link requires a file target")
		}
		if !sameVolume(targetAbs, linkAbs) {
			return errors.New("a hard link target and link must be on the same volume")
		}
	default:
		return errors.New("unsupported link type")
	}

	if samePath(targetAbs, linkAbs) {
		return errors.New("target and link path must be different")
	}
	return nil
}

func Create(target, linkPath string, typ LinkType) error {
	if err := Validate(target, linkPath, typ); err != nil {
		return err
	}
	targetAbs, _ := filepath.Abs(filepath.Clean(target))
	linkAbs, _ := filepath.Abs(filepath.Clean(linkPath))

	switch typ {
	case FileSymbolicLink:
		return createSymbolicLink(linkAbs, targetAbs, false)
	case DirectorySymbolicLink:
		return createSymbolicLink(linkAbs, targetAbs, true)
	case HardLink:
		return createHardLink(linkAbs, targetAbs)
	case Junction:
		return createJunction(linkAbs, targetAbs)
	default:
		return errors.New("unsupported link type")
	}
}

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procCreateSymbolicLinkW = kernel32.NewProc("CreateSymbolicLinkW")
	procCreateHardLinkW = kernel32.NewProc("CreateHardLinkW")
	procCreateFileW = kernel32.NewProc("CreateFileW")
	procDeviceIoControl = kernel32.NewProc("DeviceIoControl")
	procCloseHandle = kernel32.NewProc("CloseHandle")
	procGetVolumePathNameW = kernel32.NewProc("GetVolumePathNameW")
)

const (
	invalidHandleValue = ^uintptr(0)
	genericRead = 0x80000000
	genericWrite = 0x40000000
	fileShareRead = 0x00000001
	fileShareWrite = 0x00000002
	fileShareDelete = 0x00000004
	openExisting = 3
	fileFlagBackupSemantics = 0x02000000
	fileFlagOpenReparsePoint = 0x00200000
	fsctlSetReparsePoint = 0x000900A4
	ioReparseTagMountPoint = 0xA0000003
)

func createSymbolicLink(link, target string, directory bool) error {
	var flags uintptr
	if directory {
		flags = 1
	}
	r, _, err := procCreateSymbolicLinkW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(link))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(target))),
		flags,
	)
	if r == 0 {
		return winError("create symbolic link", err)
	}
	return nil
}

func createHardLink(link, target string) error {
	r, _, err := procCreateHardLinkW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(link))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(target))),
		0,
	)
	if r == 0 {
		return winError("create hard link", err)
	}
	return nil
}

func createJunction(link, target string) error {
	handle, _, err := procCreateFileW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(link))),
		genericRead|genericWrite,
		fileShareRead|fileShareWrite|fileShareDelete,
		0,
		openExisting,
		fileFlagBackupSemantics|fileFlagOpenReparsePoint,
		0,
	)
	if handle == invalidHandleValue {
		return winError("open junction path", err)
	}
	defer procCloseHandle.Call(handle)

	substitute := "\\??\\" + target
	printName := target
	sub := syscall.StringToUTF16(substitute)
	prn := syscall.StringToUTF16(printName)

	subOff := uint16(0)
	subLen := uint16(len(sub) * 2)
	printOff := subLen + 2
	printLen := uint16(len(prn) * 2)

	dataLen := uint16(8 + subLen + 2 + printLen + 2)
	buf := make([]byte, 8+int(dataLen))
	put32(buf[0:4], ioReparseTagMountPoint)
	put16(buf[4:6], dataLen)
	put16(buf[6:8], 0)
	put16(buf[8:10], subOff)
	put16(buf[10:12], subLen)
	put16(buf[12:14], printOff)
	put16(buf[14:16], printLen)

	pos := 16
	for _, u := range sub {
		put16(buf[pos:pos+2], u)
		pos += 2
	}
	pos += 2
	for _, u := range prn {
		put16(buf[pos:pos+2], u)
		pos += 2
	}
	pos += 2

	var returned uint32
	r, _, callErr := procDeviceIoControl.Call(
		handle,
		fsctlSetReparsePoint,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0, 0,
		uintptr(unsafe.Pointer(&returned)),
		0,
	)
	if r == 0 {
		return winError("create junction", callErr)
	}
	return nil
}

func put16(dst []byte, v uint16) {
	dst[0], dst[1] = byte(v), byte(v>>8)
}

func put32(dst []byte, v uint32) {
	dst[0], dst[1], dst[2], dst[3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
}

func validateLinkName(name string) error {
	if name == "" || name == "." || name == ".." {
		return errors.New("invalid link name")
	}
	if strings.HasSuffix(name, " ") || strings.HasSuffix(name, ".") {
		return errors.New("link name cannot end with a space or period")
	}
	if strings.ContainsAny(name, "<>:\"/\\\\|?*") {
		return errors.New("link name contains an invalid Windows filename character")
	}
	base := strings.TrimSuffix(strings.ToUpper(name), ".")
	switch base {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return errors.New("link name is a reserved Windows device name")
	}
	return nil
}

func samePath(a, b string) bool {
	return filepath.Clean(strings.ToLower(a)) == filepath.Clean(strings.ToLower(b))
}

func sameVolume(a, b string) bool {
	va, vb := volumeRoot(a), volumeRoot(b)
	return va != "" && vb != "" && strings.EqualFold(va, vb)
}

func volumeRoot(path string) string {
	buf := make([]uint16, 260)
	r, _, _ := procGetVolumePathNameW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func winError(op string, err error) error {
	if errno, ok := err.(syscall.Errno); ok && errno != 0 {
		return fmt.Errorf("%s: %w", op, errno)
	}
	return fmt.Errorf("%s failed", op)
}
