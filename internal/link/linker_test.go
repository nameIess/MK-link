//go:build windows

package link

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateLinkName(t *testing.T) {
	tests := []struct {
		name string
		input string
		wantErr bool
	}{
		{name: "normal", input: "my-link", wantErr: false},
		{name: "unicode", input: "मेरा-link", wantErr: false},
		{name: "dot", input: ".", wantErr: true},
		{name: "parent", input: "..", wantErr: true},
		{name: "separator", input: "bad/name", wantErr: true},
		{name: "trailing dot", input: "bad.", wantErr: true},
		{name: "trailing space", input: "bad ", wantErr: true},
		{name: "reserved", input: "CON", wantErr: true},
		{name: "reserved extension", input: "COM1.txt", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if gotErr := validateLinkName(tt.input); (gotErr != nil) != tt.wantErr {
				t.Fatalf("validateLinkName(%q) error = %v, wantErr=%v", tt.input, gotErr, tt.wantErr)
			}
		})
	}
}

func TestPlanLinkRejectsDestinationInsideDirectoryTarget(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(target, "child", "link")
	if _, err := PlanLink(target, destination, Symbolic); !errors.Is(err, ErrNestedDestination) {
		t.Fatalf("PlanLink error = %v, want %v", err, ErrNestedDestination)
	}
}

func TestPlanLinkRejectsDirectoryHardlink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(root, "link")
	if _, err := PlanLink(target, destination, Hardlink); !errors.Is(err, ErrTargetTypeMismatch) {
		t.Fatalf("PlanLink error = %v, want %v", err, ErrTargetTypeMismatch)
	}
}

func TestPlanLinkRejectsExistingDestination(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	destination := filepath.Join(root, "link")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := PlanLink(target, destination, Hardlink); !errors.Is(err, ErrDestinationExists) {
		t.Fatalf("PlanLink error = %v, want %v", err, ErrDestinationExists)
	}
}

func TestTypeDescriptions(t *testing.T) {
	t.Parallel()
	for _, typ := range []Type{Symbolic, Junction, Hardlink} {
		if strings.TrimSpace(typ.Label()) == "" || strings.TrimSpace(typ.Description()) == "" {
			t.Fatalf("type %q is missing user-facing metadata", typ)
		}
	}
}
