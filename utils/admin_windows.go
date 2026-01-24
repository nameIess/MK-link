package utils

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	procShellExecute = shell32.NewProc("ShellExecuteW")

	advapi32             = syscall.NewLazyDLL("advapi32.dll")
	procOpenProcessToken = advapi32.NewProc("OpenProcessToken")
	procGetTokenInfo     = advapi32.NewProc("GetTokenInformation")
)

const (
	TokenElevation = 20
)

// IsAdmin checks if the current process has administrative privileges.
func IsAdmin() bool {
	var token syscall.Token
	// Open the process token
	// STANDARD_RIGHTS_READ | TOKEN_QUERY = 0x20008
	h, _ := syscall.GetCurrentProcess()
	err := syscall.OpenProcessToken(h, 0x20008, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	var elevation uint32
	var cbSize uint32
	// GetTokenInformation for TokenElevation
	_, _, err = procGetTokenInfo.Call(
		uintptr(token),
		uintptr(TokenElevation),
		uintptr(unsafe.Pointer(&elevation)),
		uintptr(4),
		uintptr(unsafe.Pointer(&cbSize)),
	)

	// If the call succeeds (err == 0/nil in syscall wrapper context usually, but here err is syscall.Errno)
	// In Go syscalls, err is always non-nil if failure, but windows returns 0 on success.
	// Actually syscall.Syscall returns (r1, r2, err). If success, err is 0.

	// Check if elevation is non-zero
	return elevation != 0
}

// RunAsAdmin relaunches the current executable with administrative privileges.
func RunAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// Arguments
	args := strings.Join(os.Args[1:], " ")

	// ShellExecuteW(hwnd, "runas", exe, args, dir, showCmd)
	ret, _, err := procShellExecute.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("runas"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(exe))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(args))),
		0,
		1, // SW_SHOWNORMAL
	)

	// ShellExecute returns >32 on success
	if ret <= 32 {
		return fmt.Errorf("ShellExecute failed with code %d: %v", ret, err)
	}
	return nil
}

// EnableVirtualTerminalProcessing enables ANSI color support in Windows console
func EnableVirtualTerminalProcessing() error {
	handle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return err
	}

	var mode uint32
	err = syscall.GetConsoleMode(handle, &mode)
	if err != nil {
		return err
	}

	const ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	mode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING

	ret, _, err := syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode").Call(uintptr(handle), uintptr(mode))
	if ret == 0 {
		return err
	}
	return nil
}
