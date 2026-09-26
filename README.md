# MK-Link

A native Windows desktop utility for creating filesystem links safely.

## Link types

**Symbolic link** — points to a file or directory through a filesystem reparse point. MK-Link creates an absolute symbolic link using the Windows API.

**Junction** — redirects one directory to another directory using a Windows mount-point reparse point. Junctions are directory-only and this application restricts them to local Windows volumes.

**Hard link** — creates another directory entry for the same file data. Hard links are file-only and the target and destination must be on the same volume.

## Build

Requires Go 1.26 or newer and Windows. The application uses native Win32 controls and Windows filesystem APIs; it does not invoke a command shell to create links.

```powershell
go test ./...
go build -ldflags="-H=windowsgui" -o MK-Link.exe .
```

The window icon is the existing `ui/icon.ico` asset; no other source code from the previous implementation is used.

## Safety behavior

The application never overwrites an existing destination. It validates link names, target type compatibility, hard-link volume requirements, junction locality, and directory self-nesting before attempting creation. Filesystem APIs are called directly and user-facing errors are separated from structured diagnostic logging.
