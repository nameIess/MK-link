//go:build windows

package linker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRejectsExistingLink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	if err := os.WriteFile(target, []byte("x"), 0600); err != nil { t.Fatal(err) }
	if err := os.WriteFile(link, []byte("x"), 0600); err != nil { t.Fatal(err) }
	if err := Validate(target, link, HardLink); err == nil { t.Fatal("expected existing-link validation error") }
}

func TestValidateTypeCompatibility(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("x"), 0600); err != nil { t.Fatal(err) }
	if err := Validate(target, link, Junction); err == nil { t.Fatal("expected directory-target validation error") }
}
