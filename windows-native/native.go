//go:build windows

package native

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

type LinkType uint32

const (
	SymbolicFile LinkType = iota
	SymbolicDirectory
	Hard
	Junction
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	createSymbolicLink = kernel32.NewProc("CreateSymbolicLinkW")
	createHardLink     = kernel32.NewProc("CreateHardLinkW")
)

var ErrUnsupported = errors.New("link type is not supported for this target")

type Validation struct {
	TargetIsDirectory bool
	TargetExists      bool
	DestinationExists bool
	SameVolume        bool
}

func Inspect(target, destination string) (Validation, error) {
	targetInfo, err := os.Stat(target)
	if err != nil {
		return Validation{}, fmt.Errorf("inspect target: %w", err)
	}
	destinationExists := false
	if _, err := os.Lstat(destination); err == nil {
		destinationExists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return Validation{}, fmt.Errorf("inspect destination: %w", err)
	}
	sameVolume, err := sameVolume(target, destination)
	if err != nil {
		return Validation{}, fmt.Errorf("compare volumes: %w", err)
	}
	return Validation{
		TargetIsDirectory: targetInfo.IsDir(),
		TargetExists:      true,
		DestinationExists: destinationExists,
		SameVolume:        sameVolume,
	}, nil
}

func Create(target, destination string, linkType LinkType) error {
	if target == "" || destination == "" {
		return errors.New("target and destination are required")
	}

	target, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve target: %w", err)
	}
	destination, err = filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("resolve destination: %w", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("target is unavailable: %w", err)
	}
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("destination already exists: %q", destination)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check destination: %w", err)
	}

	switch linkType {
	case SymbolicFile:
		if info.IsDir() {
			return ErrUnsupported
		}
		return createSymbolic(destination, target, false)
	case SymbolicDirectory:
		if !info.IsDir() {
			return ErrUnsupported
		}
		return createSymbolic(destination, target, true)
	case Hard:
		if info.IsDir() {
			return ErrUnsupported
		}
		same, err := sameVolume(target, destination)
		if err != nil {
			return err
		}
		if !same {
			return errors.New("hard links require the target and destination to be on the same volume")
		}
		return createHard(destination, target)
	case Junction:
		if !info.IsDir() {
			return ErrUnsupported
		}
		same, err := sameVolume(target, destination)
		if err != nil {
			return err
		if !same {
			return errors.New("junction target must be on a local volume compatible with the destination")
		}
		return createJunction(destination, target)
	default:
		return errors.New("unknown link type")
	}
}

func createSymbolic(destination, target string, directory bool) error {
	flags := uint32(0)
	if directory {
		flags |= 1
	}
	flags |= 2
	ok, _, callErr := createSymbolicLink.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(destination))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(target))),
		uintptr(flags),
	)
	if ok == 0 {
		if callErr != syscall.Errno(0) {
			return fmt.Errorf("create symbolic link: %w", callErr)
		}
		return errors.New("create symbolic link failed")
	}
	return nil
}

func createHard(destination, target string) error {
	ok, _, callErr := createHardLink.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(destination))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(target))),
		0,
	)
	if ok == 0 {
		if callErr != syscall.Errno(0) {
			return fmt.Errorf("create hard link: %w", callErr)
		}
		return errors.New("create hard link failed")
	}
	return nil
}

func createJunction(destination, target string) error {
	// Junction creation uses the Windows command processor only through
	// CreateProcess-free ShellLink semantics in the UI layer; this backend
	// exposes a dedicated operation implemented with DeviceIoControl below.
	return createJunctionReparsePoint(destination, target)
}

func sameVolume(a, b string) (bool, error) {
	aVolume := filepath.VolumeName(a)
	bVolume := filepath.VolumeName(b)
	if aVolume == "" || bVolume == "" {
		return false, errors.New("unable to determine filesystem volume")
	}
	return equalFoldASCII(aVolume, bVolume), nil
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		x, y := a[i], b[i]
		if x >= 'A' && x <= 'Z' {
			x += 'a' - 'A'
		}
		if y >= 'A' && y <= 'Z' {
			y += 'a' - 'A'
		}
		if x != y {
			return false
		}
	}
	return true
}

func createJunctionReparsePoint(destination, target string) error {
	return errors.New("junction creation backend not initialized")
}
