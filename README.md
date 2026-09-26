# MK-Link

MK-Link is a Windows-native GUI for creating filesystem links without a command prompt. It uses Wails v2 for the desktop shell and an embedded HTML/CSS/TypeScript frontend.

## Link types

### Symbolic link

A symbolic link is a filesystem object that points to another path. MK-Link supports file and directory symbolic links. Windows permits absolute or relative symbolic-link targets; MK-Link uses absolute paths for predictable behavior.

### Junction

A junction is a Windows directory reparse point that links one directory path to another. It is directory-only and can link directories on different local volumes.

### Hard link

A hard link gives another directory entry to the same underlying file. It is file-only and the target and link must be on the same volume.

## Requirements

Windows 10/11, Go 1.25+, Node.js/npm, and Wails v2.15.0. Wails uses Microsoft WebView2 on Windows.

## Development

Install the Wails v2 CLI, then run:

    wails dev

Create a production Windows build with:

    wails build -webview2 embed -trimpath

Create an NSIS installer with:

    wails build -nsis

## Implementation

The rewrite does not reuse the previous application source. The only retained project asset is `ui/icon.ico`, copied unchanged for the new application icon.

Link creation uses Windows filesystem APIs directly. No `cmd.exe`, `mklink`, shell parsing, or user-controlled command strings are used.

Validation checks the target type, target existence, link-path collision, and hard-link volume compatibility. The optional overwrite path removes only the exact existing entry; it never recursively deletes a directory.

## License

See [License](./License).
