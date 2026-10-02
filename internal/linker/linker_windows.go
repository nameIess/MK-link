//go:build windows

package linker

import (
    "encoding/binary"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "golang.org/x/sys/windows"

    "mklink/internal/core"
)

const (
    symbolicLinkFlagDirectory = 0x1
    symbolicLinkFlagAllowUnprivileged = 0x2
    fsctlSetReparsePoint = 0x000900A4
    ioReparseTagMountPoint = 0xA0000003

    genericWrite = 0x40000000
    fileShareRead = 0x00000001
    fileShareWrite = 0x00000002
    fileShareDelete = 0x00000004
    openExisting = 3
    fileFlagOpenReparsePoint = 0x00200000
    fileFlagBackupSemantics = 0x02000000

    errorAccessDenied = windows.Errno(5)
    errorPrivilegeNotHeld = windows.Errno(1314)
)

func Create(req core.ValidatedCreateRequest) error {
    switch req.LinkType {
    case core.LinkTypeSymbolic:
        return createSymbolic(req.TargetPath, req.LinkPath)
    case core.LinkTypeHard:
        return createHard(req.TargetPath, req.LinkPath)
    case core.LinkTypeJunction:
        return createJunction(req.TargetPath, req.LinkPath)
    default:
        return fmt.Errorf("unsupported link type")
    }
}

func createSymbolic(target, link string) error {
    targetPtr, err := windows.UTF16PtrFromString(longPath(target))
    if err != nil { return err }
    linkPtr, err := windows.UTF16PtrFromString(longPath(link))
    if err != nil { return err }

    flags := uint32(symbolicLinkFlagAllowUnprivileged)
    info, err := os.Stat(target)
    if err != nil { return fmt.Errorf("inspect symbolic link target: %w", err) }
    if info.IsDir() { flags |= symbolicLinkFlagDirectory }

    if err := windows.CreateSymbolicLink(linkPtr, targetPtr, flags); err != nil {
        return fmt.Errorf("create symbolic link: %w", err)
    }
    return nil
}

func createHard(target, link string) error {
    targetPtr, err := windows.UTF16PtrFromString(longPath(target))
    if err != nil { return err }
    linkPtr, err := windows.UTF16PtrFromString(longPath(link))
    if err != nil { return err }

    if err := windows.CreateHardLink(linkPtr, targetPtr, 0); err != nil {
        return fmt.Errorf("create hard link: %w", err)
    }
    return nil
}

func createJunction(target, link string) error {
    if strings.HasPrefix(target, "\\") {
        return fmt.Errorf("junctions cannot target a UNC or network path")
    }

    if err := windows.CreateDirectory(mustUTF16(longPath(link)), nil); err != nil {
        return fmt.Errorf("create junction placeholder: %w", err)
    }

    keep := false
    defer func() {
        if !keep { _ = os.Remove(link) }
    }()

    handle, err := windows.CreateFile(
        mustUTF16(longPath(link)),
        genericWrite,
        fileShareRead|fileShareWrite|fileShareDelete,
        nil,
        openExisting,
        fileFlagOpenReparsePoint|fileFlagBackupSemantics,
        0,
    )
    if err != nil { return fmt.Errorf("open junction placeholder: %w", err) }
    defer windows.CloseHandle(handle)

    buffer, err := junctionBuffer(target)
    if err != nil { return err }

    var returned uint32
    if err := windows.DeviceIoControl(handle, fsctlSetReparsePoint, &buffer[0], uint32(len(buffer)), nil, 0, &returned, nil); err != nil {
        return fmt.Errorf("set junction reparse point: %w", err)
    }
    keep = true
    return nil
}

func junctionBuffer(target string) ([]byte, error) {
    clean := filepath.Clean(strings.TrimPrefix(target, `\\?\`))
    if len(clean) < 3 || clean[1] != ':' || clean[2] != '\\' {
        return nil, fmt.Errorf("junction targets must be local absolute paths")
    }

    substitute := `\??\` + clean
    printName := clean
    sub := windows.StringToUTF16(substitute)
    print := windows.StringToUTF16(printName)

    subBytes := make([]byte, len(sub)*2)
    printBytes := make([]byte, len(print)*2)
    for i, v := range sub { binary.LittleEndian.PutUint16(subBytes[i*2:], v) }
    for i, v := range print { binary.LittleEndian.PutUint16(printBytes[i*2:], v) }

    dataLen := 8 + len(subBytes) + 2 + len(printBytes) + 2
    buf := make([]byte, 8+dataLen)

    binary.LittleEndian.PutUint32(buf[0:], ioReparseTagMountPoint)
    binary.LittleEndian.PutUint16(buf[4:], uint16(dataLen))
    binary.LittleEndian.PutUint16(buf[6:], 0)

    base := 8
    binary.LittleEndian.PutUint16(buf[base+0:], 0)
    binary.LittleEndian.PutUint16(buf[base+2:], uint16(len(subBytes)))
    binary.LittleEndian.PutUint16(buf[base+4:], uint16(len(subBytes)+2))
    binary.LittleEndian.PutUint16(buf[base+6:], uint16(len(printBytes)))

    cursor := base + 8
    copy(buf[cursor:], subBytes)
    cursor += len(subBytes) + 2
    copy(buf[cursor:], printBytes)
    return buf, nil
}

func mustUTF16(path string) *uint16 {
    p, _ := windows.UTF16PtrFromString(path)
    return p
}

func longPath(path string) string {
    if strings.HasPrefix(path, `\\?\`) || len(path) < 248 {
        return path
    }
    if strings.HasPrefix(path, `\\`) {
        return `\\?\UNC\` + strings.TrimPrefix(path, `\\`)
    }
    return `\\?\` + path
}

func NeedsElevation(err error) bool {
    if err == nil { return false }
    return errors.Is(err, errorAccessDenied) ||
        errors.Is(err, errorPrivilegeNotHeld) ||
        strings.Contains(strings.ToLower(err.Error()), "access is denied") ||
        strings.Contains(strings.ToLower(err.Error()), "privilege")
}

func UserFacingError(err error) error {
    if err == nil { return nil }
    if NeedsElevation(err) {
        return fmt.Errorf("Windows denied this operation. Try running MK-Link as administrator; symbolic links may also work without elevation when Windows Developer Mode is enabled")
    }
    if errors.Is(err, os.ErrExist) {
        return core.ErrDestinationExists
    }
    return err
}

func IsElevated() bool {
    return windows.GetCurrentProcessToken().IsElevated()
}

func OpenExplorer(path string) error {
    verb, err := windows.UTF16PtrFromString("open")
    if err != nil { return err }
    target, err := windows.UTF16PtrFromString(longPath(path))
    if err != nil { return err }
    return windows.ShellExecute(0, verb, target, nil, nil, 1)
}
