# MK-Link Native Windows UI

This directory contains the new native Windows implementation of MK-Link.

The implementation is intentionally independent of the legacy TUI source. The only asset carried forward from the existing application is `ui/icon.ico`, reused as the executable icon.

## Link types

- **Symbolic link** — a filesystem object that points to another file or directory. Windows supports absolute and relative symbolic links; creation may require elevation unless the system/user policy permits unprivileged creation.
- **Junction** — a Windows directory reparse point that redirects one directory path to another. Junctions are directory-only and can target another local volume, but not a mapped network volume.
- **Hard link** — an additional directory entry for the same file data. Hard links are file-only and must reference a file on the same volume.

The UI exposes only valid link types for the selected target and validates the destination before creation.
