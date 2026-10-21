# MK-Link v3.0 🚀

> A fast, focused Windows utility for creating filesystem links through a guided native interface.

## Highlights

- Create symbolic links for files and directories.
- Create directory junctions for local Windows folders.
- Create hard links for files on the same volume.
- Pick files and folders with native Windows dialogs.
- Drag and drop paths from Windows Explorer.
- Validate paths and detect destination conflicts before changes are made.
- Review operations before filesystem mutation.
- Harden path validation, including destination paths that resolve through symlink ancestors.
- Improve frontend handling of overlapping asynchronous picker and validation operations.
- Improve keyboard focus visibility, reduced-motion support, and the About dialog.
- Simplify Windows icon preparation to use the checked-in ICO asset without requiring ImageMagick.
- Reduce duplicate Windows CI runs by cancelling superseded builds.

## Download

Download **`mk-link.exe`** from the GitHub Release assets and run it on Windows.

## Requirements

- Windows 10 or Windows 11.
- Microsoft Edge WebView2 runtime.
- Some link operations may require administrator privileges depending on Windows policy and the target filesystem.

## Build and verification

The Windows CI workflow checks frontend compilation, Go tests, Go vet, and the production Wails build. Use the published v3.0 release asset once the release workflow has completed successfully.

## Notes

This release is intended for Windows x64. Path-based filesystem operations still have a narrow time-of-check/time-of-use race under adversarial concurrent filesystem changes; do not treat validation as a security boundary against a process actively changing the filesystem.
