# MK-Link

A Windows-native GUI for creating filesystem links without a command shell.

## Link types

- **Symbolic link** — a filesystem entry that redirects access to another path. A file symlink targets a file; a directory symlink targets a directory. Symlinks can cross volumes.
- **Junction** — a Windows directory reparse point that redirects a directory path. Junctions are for directories and the target/link must be on the same volume.
- **Hard link** — a second directory entry for the same file data. Hard links are for files only and must remain on the same volume. Removing one name does not remove the file data while another hard-link name remains.

MK-Link validates the target, destination, link name, existing paths, type compatibility, and same-volume requirements before calling native Windows APIs.

## Security model

The application never invokes `cmd.exe`, PowerShell, or shell commands to create links. User-controlled paths are passed directly to Windows APIs. Existing files are never overwritten.

Creating symbolic links may require the Windows symbolic-link privilege unless Developer Mode or an equivalent policy permits unprivileged creation. MK-Link reports the Windows error rather than silently elevating the process.

## Build

```powershell
go test ./...
go vet ./...
go build -trimpath -ldflags="-s -w -H=windowsgui" -o mklink.exe .
```

The GUI loads the repository's supplied `ui/icon.ico` at runtime. Package the `ui` directory alongside `mklink.exe`.

## Supported systems

Windows 10 and Windows 11, 64-bit.
