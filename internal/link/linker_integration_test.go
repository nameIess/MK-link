//go:build integration && windows

package link

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateLinks(t *testing.T) {
	root := t.TempDir()
	targetFile := filepath.Join(root, "target.txt")
	targetDir := filepath.Join(root, "target-dir")
	if err := os.WriteFile(targetFile, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		typ Type
		target string
		destination string
	}{
		{name: "hardlink", typ: Hardlink, target: targetFile, destination: filepath.Join(root, "hard.txt")},
		{name: "symbolic-file", typ: Symbolic, target: targetFile, destination: filepath.Join(root, "symbolic-file.txt")},
		{name: "symbolic-directory", typ: Symbolic, target: targetDir, destination: filepath.Join(root, "symbolic-dir")},
		{name: "junction", typ: Junction, target: targetDir, destination: filepath.Join(root, "junction")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plan, err := PlanLink(tt.target, tt.destination, tt.typ)
			if err != nil {
				t.Fatalf("PlanLink: %v", err)
			}
			if err := Create(plan); err != nil {
				if tt.typ == Symbolic && isPrivilegeError(err) {
					t.Skipf("symbolic links are disabled in this runner: %v", err)
				}
				t.Fatalf("Create: %v", err)
			}
			info, err := os.Lstat(tt.destination)
			if err != nil {
				t.Fatalf("Lstat destination: %v", err)
			}
			switch tt.typ {
			case Junction, Symbolic:
				if info.Mode()&os.ModeSymlink == 0 && tt.typ == Symbolic {
					t.Fatal("symbolic link destination is not a symlink")
				}
			case Hardlink:
				infoTarget, err := os.Stat(tt.target)
				if err != nil {
					t.Fatal(err)
				}
				if info.Size() != infoTarget.Size() {
					t.Fatalf("hardlink size = %d, target size = %d", info.Size(), infoTarget.Size())
				}
			}
		})
	}
}

func isPrivilegeError(err error) bool {
	var errno windows.Errno
	return errors.As(err, &errno) && errno == windows.ERROR_PRIVILEGE_NOT_HELD
}
