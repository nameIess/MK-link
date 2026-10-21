package core

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "unicode"
)

var (
    ErrInvalidPath = errors.New("the selected path is invalid")
    ErrDestinationExists = errors.New("the destination already exists")
    ErrInvalidName = errors.New("the link name is invalid")
)

func AvailableLinkTypes(target string) []LinkTypeInfo {
    info, err := InspectPath(target)
    if err != nil { return nil }
    if info.IsFile {
        return []LinkTypeInfo{
            {ID: LinkTypeSymbolic, Name: "Symbolic Link", Summary: "A flexible filesystem reference to another file.", BestFor: "Development workflows, compatibility paths, and references that may cross volumes.", TargetKind: "Files", Description: "Stores a reference to the target path. It can point to a file on another volume or a network location."},
            {ID: LinkTypeHard, Name: "Hard Link", Summary: "Another directory entry for the same file data.", BestFor: "Multiple filenames for the same file without duplicating its contents.", TargetKind: "Files on the same volume", Description: "Both names refer to the same underlying file record. Removing one name does not remove the data while another hard link remains."},
        }
    }
    if info.IsDir {
        return []LinkTypeInfo{
            {ID: LinkTypeSymbolic, Name: "Symbolic Link", Summary: "A flexible filesystem reference to another directory.", BestFor: "Redirecting folders, development environments, and paths that may cross volumes.", TargetKind: "Files & directories", Description: "Stores a reference to the target directory and can point across volumes or to a UNC path."},
            {ID: LinkTypeJunction, Name: "Junction", Summary: "A Windows directory reparse point.", BestFor: "Redirecting large local folders while applications keep using the original path.", TargetKind: "Local directories", Description: "A Windows mount-point reparse point. Junctions are directory-only and are not intended for remote shares."},
        }
    }
    return nil
}

func InspectPath(path string) (PathInfo, error) {
    clean, err := NormalizeExistingPath(path)
    if err != nil { return PathInfo{}, err }
    info, err := os.Lstat(clean)
    if err != nil { return PathInfo{}, err }
    return PathInfo{Path: clean, Exists: true, IsFile: !info.IsDir(), IsDir: info.IsDir(), IsReparse: isReparsePoint(clean, info), Size: info.Size()}, nil
}

func NormalizeExistingPath(path string) (string, error) {
    clean, err := normalizePath(path)
    if err != nil { return "", err }
    if _, err := os.Lstat(clean); err != nil { return "", fmt.Errorf("%w: %s", ErrInvalidPath, clean) }
    return clean, nil
}

func ValidateDestination(directory, name string, linkType LinkType) (ValidationResult, error) {
    dir, err := NormalizeExistingPath(directory)
    if err != nil { return ValidationResult{}, err }
    info, err := os.Stat(dir)
    if err != nil || !info.IsDir() { return ValidationResult{}, fmt.Errorf("destination must be an existing directory") }
    cleanName, err := ValidateLinkName(name)
    if err != nil { return ValidationResult{}, err }
    switch linkType {
    case LinkTypeSymbolic, LinkTypeHard, LinkTypeJunction:
    default:
        return ValidationResult{}, fmt.Errorf("unsupported link type")
    }
    linkPath := filepath.Join(dir, cleanName)
    _, statErr := os.Lstat(linkPath)
    if statErr == nil {
        return ValidationResult{Valid: false, Conflict: true, LinkPath: linkPath, Message: "A file, folder, or link already exists at the destination."}, nil
    }
    if !os.IsNotExist(statErr) { return ValidationResult{}, fmt.Errorf("cannot inspect destination: %w", statErr) }
    return ValidationResult{Valid: true, LinkPath: linkPath, Message: "Destination is available."}, nil
}

func ValidateCreateRequest(req CreateLinkRequest) (ValidatedCreateRequest, error) {
    target, err := InspectPath(req.TargetPath)
    if err != nil { return ValidatedCreateRequest{}, err }
    if target.IsReparse { return ValidatedCreateRequest{}, fmt.Errorf("the selected target is a reparse point; select its real file or directory instead") }

    dir, err := NormalizeExistingPath(req.DestinationDir)
    if err != nil { return ValidatedCreateRequest{}, err }
    dirInfo, err := os.Stat(dir)
    if err != nil || !dirInfo.IsDir() { return ValidatedCreateRequest{}, fmt.Errorf("destination must be an existing directory") }
    dirEntry, err := os.Lstat(dir)
    if err != nil { return ValidatedCreateRequest{}, fmt.Errorf("cannot inspect destination: %w", err) }
    if isReparsePoint(dir, dirEntry) { return ValidatedCreateRequest{}, fmt.Errorf("destination directory cannot be a reparse point") }

    name, err := ValidateLinkName(req.LinkName)
    if err != nil { return ValidatedCreateRequest{}, err }

    switch req.LinkType {
    case LinkTypeSymbolic:
    case LinkTypeHard:
        if !target.IsFile { return ValidatedCreateRequest{}, fmt.Errorf("hard links can only target files") }
    case LinkTypeJunction:
        if !target.IsDir { return ValidatedCreateRequest{}, fmt.Errorf("junctions can only target directories") }
    default:
        return ValidatedCreateRequest{}, fmt.Errorf("unsupported link type")
    }

    linkPath := filepath.Join(dir, name)
    if _, err := os.Lstat(linkPath); err == nil {
        return ValidatedCreateRequest{}, ErrDestinationExists
    } else if !os.IsNotExist(err) {
        return ValidatedCreateRequest{}, fmt.Errorf("cannot inspect destination: %w", err)
    }

    // Resolve existing paths before checking containment. Lexical path checks
    // miss destinations routed through an ancestor junction or symlink.
    targetAbs, err := filepath.Abs(target.Path)
    if err != nil { return ValidatedCreateRequest{}, fmt.Errorf("resolve target path: %w", err) }
    targetResolved, err := filepath.EvalSymlinks(targetAbs)
    if err != nil { return ValidatedCreateRequest{}, fmt.Errorf("resolve target path: %w", err) }
    dirAbs, err := filepath.Abs(dir)
    if err != nil { return ValidatedCreateRequest{}, fmt.Errorf("resolve destination path: %w", err) }
    dirResolved, err := filepath.EvalSymlinks(dirAbs)
    if err != nil { return ValidatedCreateRequest{}, fmt.Errorf("resolve destination path: %w", err) }

    resolvedLinkPath := filepath.Join(dirResolved, name)
    if strings.EqualFold(filepath.Clean(targetResolved), filepath.Clean(resolvedLinkPath)) {
        return ValidatedCreateRequest{}, fmt.Errorf("the link destination cannot be the target itself")
    }
    if target.IsDir && isPathInside(targetResolved, dirResolved) {
        return ValidatedCreateRequest{}, fmt.Errorf("the link destination cannot be inside the directory it targets; that would create a recursive link")
    }

    return ValidatedCreateRequest{TargetPath: target.Path, LinkPath: linkPath, LinkType: req.LinkType}, nil
}

func isReparsePoint(path string, info os.FileInfo) bool {
    if info.Mode()&os.ModeSymlink != 0 {
        return true
    }
    _, err := os.Readlink(path)
    return err == nil
}

func ValidateLinkName(name string) (string, error) {
    if name == "" || name == "." || name == ".." || len(name) > 240 { return "", ErrInvalidName }
    if filepath.Base(name) != name || strings.ContainsAny(name, "/\\") { return "", fmt.Errorf("%w: use a single filename without path separators", ErrInvalidName) }
    if strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") { return "", fmt.Errorf("%w: Windows names cannot end with a dot or space", ErrInvalidName) }
    if strings.Contains(name, ":") { return "", fmt.Errorf("%w: alternate data stream syntax is not allowed", ErrInvalidName) }
    for _, r := range name {
        if unicode.IsControl(r) || strings.ContainsRune(`<>:"|?*`, r) { return "", fmt.Errorf("%w: contains a Windows-invalid character", ErrInvalidName) }
    }
    stem := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
    switch stem {
    case "CON","PRN","AUX","NUL","COM1","COM2","COM3","COM4","COM5","COM6","COM7","COM8","COM9","LPT1","LPT2","LPT3","LPT4","LPT5","LPT6","LPT7","LPT8","LPT9":
        return "", fmt.Errorf("%w: reserved Windows device name", ErrInvalidName)
    }
    return name, nil
}

func isPathInside(parent, child string) bool {
    parent = filepath.Clean(parent)
    child = filepath.Clean(child)
    rel, err := filepath.Rel(parent, child)
    if err != nil { return false }
    if rel == "." { return true }
    if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) { return false }
    return !filepath.IsAbs(rel)
}

func normalizePath(path string) (string, error) {
    path = strings.Trim(strings.TrimSpace(path), `"'`)
    if path == "" { return "", ErrInvalidPath }
    abs, err := filepath.Abs(path)
    if err != nil { return "", err }
    clean := filepath.Clean(abs)
    if clean == "." { return "", ErrInvalidPath }
    return clean, nil
}
