package logic

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type LinkType int

const (
	LinkTypeDirSymlink LinkType = iota
	LinkTypeFileSymlink
	LinkTypeHardLink
	LinkTypeJunction
)

// CreateLink creates the requested link type.
func CreateLink(linkType LinkType, targetPath, linkPath string) error {
	// Ensure target exists for safety (though symlinks can be dangling, usually we want valid ones)
	// The batch script enforced existence, so we will too.
	// However, os.Symlink doesn't require it. We'll assume validation happened before.

	// Ensure parent dir of linkPath exists
	if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	switch linkType {
	case LinkTypeDirSymlink:
		// os.Symlink arguments are (target, link)
		// On Windows, if target is relative, it's relative to the link location.
		// If we use absolute paths (which we do), it works fine.
		return os.Symlink(targetPath, linkPath)

	case LinkTypeFileSymlink:
		return os.Symlink(targetPath, linkPath)

	case LinkTypeHardLink:
		return os.Link(targetPath, linkPath)

	case LinkTypeJunction:
		return createJunction(targetPath, linkPath)
	}

	return fmt.Errorf("unknown link type")
}

func createJunction(target, link string) error {
	// Native junction creation is complex. Using mklink /j wrapper.
	// mklink is a cmd internal command.
	cmd := exec.Command("cmd", "/c", "mklink", "/j", link, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mklink failed: %v, output: %s", err, string(output))
	}
	return nil
}
