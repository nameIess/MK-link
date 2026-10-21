package core

import (
    "os"
    "path/filepath"
    "testing"
)

func TestValidateLinkName(t *testing.T) {
    for _, name := range []string{"MyLink","project-data","file.txt","backup_01"} {
        if _, err := ValidateLinkName(name); err != nil { t.Fatalf("expected %q to be valid: %v", name, err) }
    }
    for _, name := range []string{"",".","..","a/b","a\\b","CON","NUL.txt","CON.foo.txt","LPT1.backup","foo:","foo.","foo ","foo*bar"} {
        if _, err := ValidateLinkName(name); err == nil { t.Fatalf("expected %q to be invalid", name) }
    }
}

func TestIsPathInside(t *testing.T) {
    parent := filepath.Join(string(filepath.Separator), "Projects", "App")
    if !isPathInside(parent, filepath.Join(parent, "links")) { t.Fatal("expected child path to be inside") }
    if isPathInside(parent, filepath.Join(string(filepath.Separator), "Projects", "Other")) { t.Fatal("expected sibling path to be outside") }
    if isPathInside(parent, filepath.Join(string(filepath.Separator), "Projects", "AppData")) { t.Fatal("prefix-only match must not count as inside") }
}

func TestValidateCreateRequestRejectsDestinationInsideTargetViaSymlinkAncestor(t *testing.T) {
    root := t.TempDir()
    target := filepath.Join(root, "target")
    if err := os.Mkdir(target, 0o755); err != nil { t.Fatal(err) }
    child := filepath.Join(target, "child")
    if err := os.Mkdir(child, 0o755); err != nil { t.Fatal(err) }

    alias := filepath.Join(root, "alias")
    if err := os.Symlink(target, alias); err != nil {
        t.Skipf("symlink creation unavailable on this host: %v", err)
    }

    _, err := ValidateCreateRequest(CreateLinkRequest{
        TargetPath: target,
        DestinationDir: filepath.Join(alias, "child"),
        LinkName: "recursive-link",
        LinkType: LinkTypeSymbolic,
    })
    if err == nil {
        t.Fatal("expected destination reached through a symlink ancestor to be rejected as recursive")
    }
}
