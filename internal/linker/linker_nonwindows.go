//go:build !windows

package linker

import (
    "errors"
    "mklink/internal/core"
)

func Create(core.ValidatedCreateRequest) error { return errors.New("MK-Link supports Windows only") }
func NeedsElevation(error) bool { return false }
func UserFacingError(err error) error { return err }
func IsElevated() bool { return false }
func OpenExplorer(string) error { return errors.New("MK-Link supports Windows only") }
