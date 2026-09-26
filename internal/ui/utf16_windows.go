//go:build windows

package ui

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func utf16Ptr(text string) (*uint16, error) {
	ptr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return nil, fmt.Errorf("encode Windows text: %w", err)
	}
	return ptr, nil
}
