//go:build windows

package link

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Type string

const (
	Symbolic Type = "symbolic"
	Junction Type = "junction"
	Hardlink Type = "hardlink"
)

type Plan struct {
	Target string
	Destination string
	Type Type
}

var (
	ErrTargetNotFound = errors.New("target does not exist")
	ErrDestinationExists = errors.New("destination already exists")
	ErrInvalidLinkName = errors.New("invalid link name")
	ErrTargetTypeMismatch = errors.New("link type is incompatible with target")
	ErrDifferentVolume = errors.New("hard links require the same volume")
	ErrNetworkJunction = errors.New("junctions require local volumes")
	ErrNestedDestination = errors.New("destination folder cannot be inside the target directory")
)

const (
	symbolicLinkFlagDirectory = 0x00000001
	symbolicLinkFlagAllowUnprivileged = 0x00000002
	genericReadAccess = 0x80000000
	genericWriteAccess = 0x40000000
	fileShareRead = 0x00000001
	fileShareWrite = 0x00000002
	fileShareDelete = 0x00000004
	openExisting = 3
	fileFlagBackupSemantics = 0x02000000
	fileFlagOpenReparsePoint = 0x00200000
	fsctlSetReparsePoint = 0x000900A4
	ioReparseTagMountPoint = 0xA0000003
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	createSymbolicLinkW = kernel32.NewProc("CreateSymbolicLinkW")
	createHardLinkW = kernel32.NewProc("CreateHardLinkW")
	deviceIoControl = kernel32.NewProc("DeviceIoControl")
)

func (t Type) Label() string {
	switch t {
	case Symbolic:
		return "Symbolic link"
	case Junction:
		return "Junction"
	case Hardlink:
		return "Hard link"
	default:
		return "Unknown"
	}
}

func (t Type) Description() string {
	switch t {
	case Symbolic:
		return "Points to a file or directory. This app creates an absolute Windows symbolic link."
	case Junction:
		return "Redirects a directory through a Windows mount-point reparse point. Local volumes only."
	case Hardlink:
		return "Adds another name for the same file data. Files only, and both paths must share a volume."
	default:
		return ""
	}
}

func PlanLink(target, destination string, linkType Type) (Plan, error) {
	target, err := absoluteCleanPath(target)
	if err != nil { return Plan{}, fmt.Errorf("resolve target path: %w", err) }
	destination, err = absoluteCleanPath(destination)
	if err != nil { return Plan{}, fmt.Errorf("resolve destination path: %w", err) }
	info, err := os.Stat(target)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { return Plan{}, ErrTargetNotFound }
		return Plan{}, fmt.Errorf("inspect target: %w", err)
	}
	if _, err := os.Lstat(destination); err == nil {
		return Plan{}, ErrDestinationExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return Plan{}, fmt.Errorf("inspect destination: %w", err)
	}
	parent := filepath.Dir(destination)
	parentInfo, err := os.Stat(parent)
	if err != nil { return Plan{}, fmt.Errorf("inspect destination folder: %w", err) }
	if !parentInfo.IsDir() { return Plan{}, errors.New("destination folder is not a directory") }
	if err := validateLinkName(filepath.Base(destination)); err != nil { return Plan{}, err }
	isDirectory := info.IsDir()
	switch linkType {
	case Symbolic:
	case Hardlink:
		if isDirectory { return Plan{}, ErrTargetTypeMismatch }
		same, err := sameVolume(target, destination)
		if err != nil { return Plan{}, fmt.Errorf("compare volumes: %w", err) }
		if !same { return Plan{}, ErrDifferentVolume }
	case Junction:
		if !isDirectory { return Plan{}, ErrTargetTypeMismatch }
		if !isLocalVolumePath(target) || !isLocalVolumePath(destination) { return Plan{}, ErrNetworkJunction }
	default:
		return Plan{}, errors.New("unknown link type")
	}
	if isDirectory && pathInside(target, parent) { return Plan{}, ErrNestedDestination }
	return Plan{Target: target, Destination: destination, Type: linkType}, nil
}

func Create(p Plan) error {
	if p.Target == "" || p.Destination == "" { return errors.New("target and destination are required") }
	switch p.Type {
	case Symbolic:
		isDirectory, err := isDir(p.Target); if err != nil { return err }
		return createSymbolic(p.Target, p.Destination, isDirectory)
	case Hardlink:
		return createHardlink(p.Target, p.Destination)
	case Junction:
		return createJunction(p.Target, p.Destination)
	default:
		return errors.New("unknown link type")
	}
}

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil { return false, fmt.Errorf("inspect target: %w", err) }
	return info.IsDir(), nil
}

func absoluteCleanPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" { return "", errors.New("path is required") }
	absolute, err := filepath.Abs(path)
	if err != nil { return "", fmt.Errorf("make path absolute: %w", err) }
	return filepath.Clean(absolute), nil
}

func validateLinkName(name string) error {
	if name == "" || name == "." || name == ".." { return ErrInvalidLinkName }
	if len(utf16.Encode([]rune(name))) > 255 { return ErrInvalidLinkName }
	if strings.HasSuffix(name, " ") || strings.HasSuffix(name, ".") { return ErrInvalidLinkName }
	if strings.ContainsAny(name, `/\\:*?"<>|`) { return ErrInvalidLinkName }
	baseUpper := strings.ToUpper(strings.TrimRight(name, " ."))
	for _, reserved := range []string{"CON","PRN","AUX","NUL","COM1","COM2","COM3","COM4","COM5","COM6","COM7","COM8","COM9","LPT1","LPT2","LPT3","LPT4","LPT5","LPT6","LPT7","LPT8","LPT9"} {
		if baseUpper == reserved || strings.HasPrefix(baseUpper, reserved+".") { return ErrInvalidLinkName }
	}
	return nil
}

func pathInside(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	if err != nil { return false }
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

func sameVolume(a, b string) (bool, error) {
	left := strings.ToUpper(filepath.VolumeName(a)); right := strings.ToUpper(filepath.VolumeName(b))
	if left == "" || right == "" { return false, errors.New("unable to determine filesystem volume") }
	return left == right, nil
}

func isLocalVolumePath(path string) bool {
	volume := filepath.VolumeName(path); return len(volume) == 2 && volume[1] == ':'
}

func createSymbolic(target, destination string, directory bool) error {
	targetPtr, err := windows.UTF16PtrFromString(target); if err != nil { return fmt.Errorf("encode target path: %w", err) }
	destinationPtr, err := windows.UTF16PtrFromString(destination); if err != nil { return fmt.Errorf("encode destination path: %w", err) }
	flags := uint32(symbolicLinkFlagAllowUnprivileged); if directory { flags |= symbolicLinkFlagDirectory }
	r1, _, callErr := createSymbolicLinkW.Call(uintptr(unsafe.Pointer(destinationPtr)), uintptr(unsafe.Pointer(targetPtr)), uintptr(flags))
	if r1 == 0 { return fmt.Errorf("create symbolic link: %w", callErr) }
	return nil
}

func createHardlink(target, destination string) error {
	targetPtr, err := windows.UTF16PtrFromString(target); if err != nil { return fmt.Errorf("encode target path: %w", err) }
	destinationPtr, err := windows.UTF16PtrFromString(destination); if err != nil { return fmt.Errorf("encode destination path: %w", err) }
	r1, _, callErr := createHardLinkW.Call(uintptr(unsafe.Pointer(destinationPtr)), uintptr(unsafe.Pointer(targetPtr)), 0)
	if r1 == 0 { return fmt.Errorf("create hard link: %w", callErr) }
	return nil
}

func createJunction(target, destination string) error {
	if !isLocalVolumePath(target) || !isLocalVolumePath(destination) { return ErrNetworkJunction }
	targetUTF16, err := windows.UTF16FromString(`\??\` + target); if err != nil { return fmt.Errorf("encode junction target: %w", err) }
	printUTF16, err := windows.UTF16FromString(target); if err != nil { return fmt.Errorf("encode junction display name: %w", err) }
	targetUTF16 = targetUTF16[:len(targetUTF16)-1]; printUTF16 = printUTF16[:len(printUTF16)-1]
	subBytes := len(targetUTF16) * 2; printBytes := len(printUTF16) * 2; dataLen := 8 + subBytes + printBytes
	if dataLen > 0xFFFF { return errors.New("junction target path is too long") }
	buffer := make([]byte, 16+subBytes+printBytes)
	putUint32LE(buffer[0:4], ioReparseTagMountPoint); putUint16LE(buffer[4:6], uint16(dataLen)); putUint16LE(buffer[6:8], 0)
	putUint16LE(buffer[8:10], 0); putUint16LE(buffer[10:12], uint16(subBytes)); putUint16LE(buffer[12:14], uint16(subBytes)); putUint16LE(buffer[14:16], uint16(printBytes))
	appendUTF16LE(buffer[16:], targetUTF16); appendUTF16LE(buffer[16+subBytes:], printUTF16)
	if err := os.Mkdir(destination, 0o755); err != nil { return fmt.Errorf("create junction directory: %w", err) }
	created := false; defer func(){ if !created { _ = os.Remove(destination) } }()
	handle, err := openReparseDirectory(destination); if err != nil { return err }; defer windows.CloseHandle(handle)
	var returned uint32
	r1, _, callErr := deviceIoControl.Call(uintptr(handle), uintptr(fsctlSetReparsePoint), uintptr(unsafe.Pointer(&buffer[0])), uintptr(uint32(len(buffer))), 0, 0, uintptr(unsafe.Pointer(&returned)), 0)
	runtime.KeepAlive(buffer)
	if r1 == 0 { return fmt.Errorf("set junction reparse point: %w", callErr) }
	created = true; return nil
}

func openReparseDirectory(path string) (windows.Handle, error) {
	ptr, err := windows.UTF16PtrFromString(path); if err != nil { return 0, fmt.Errorf("encode junction path: %w", err) }
	handle, err := windows.CreateFile(ptr, genericReadAccess|genericWriteAccess, fileShareRead|fileShareWrite|fileShareDelete, nil, openExisting, fileFlagBackupSemantics|fileFlagOpenReparsePoint, 0)
	if err != nil { return 0, fmt.Errorf("open junction directory: %w", err) }
	return handle, nil
}

func putUint16LE(dst []byte, value uint16) { dst[0] = byte(value); dst[1] = byte(value >> 8) }
func putUint32LE(dst []byte, value uint32) { putUint16LE(dst[:2], uint16(value)); putUint16LE(dst[2:4], uint16(value>>16)) }
func appendUTF16LE(dst []byte, value []uint16) { for i, char := range value { putUint16LE(dst[i*2:], char) } }