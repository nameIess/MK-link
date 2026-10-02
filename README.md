# MK-Link

MK-Link is a Windows desktop utility for creating and managing filesystem links through a guided native interface.

It is built for people who need the behavior of Windows links without having to remember the command-line syntax or manually validate every path.

## What it does

MK-Link supports three Windows link types:

| Link type | Target | Typical use |
| --- | --- | --- |
| **Symbolic link** | File or directory | Flexible references, development paths, compatibility paths |
| **Hard link** | File only | Multiple filenames pointing to the same file data |
| **Junction** | Directory only | Redirecting local directories while keeping the original path |

Before making a filesystem change, MK-Link validates the requested operation and shows a review step. Existing destinations are not overwritten automatically.

## Features

- Guided workflow for choosing a target, link name, and link type
- Native Windows file and folder pickers
- Drag and drop from Windows Explorer
- Pre-flight validation and destination conflict detection
- Review before filesystem mutation
- Windows-native filesystem APIs rather than shelling out to `cmd.exe`, `mklink.exe`, or similar commands
- Clear explanations and restrictions for each link type
- Dark desktop interface designed for keyboard and mouse use
- Single-instance application behavior
- Windows executable produced automatically by GitHub Actions

## Security-focused behavior

Filesystem paths are treated as untrusted input.

- User-controlled paths are not inserted into shell commands.
- Link names reject path separators, alternate data streams, invalid Windows characters, reserved device names, and trailing dots or spaces.
- Existing destinations are rejected instead of being overwritten automatically.
- The backend validates the target type again immediately before creating a link.
- Symbolic links use the Windows native API and request the unprivileged-create flag when supported.
- Junctions use the Windows reparse-point API.
- UNC/network targets are rejected where the selected link type requires a local target.
- Production WebView context menus and developer tools remain disabled.

## Technology

- **Backend:** Go
- **Desktop framework:** Wails v2.15
- **Frontend:** React + TypeScript
- **Styling:** Tailwind CSS
- **Build:** Go 1.26 + Node.js 22
- **Platform:** Windows 10/11

Wails uses the Windows WebView2 runtime for the application UI.

## Project layout

```text
.
├── frontend/              # React/TypeScript application
├── internal/
│   ├── core/              # Validation and link-type rules
│   └── linker/            # Windows and non-Windows linker implementations
├── ui/
│   └── icon.svg           # Source application icon
├── .github/workflows/     # CI and release automation
├── app.go                 # Wails application bindings
├── main.go                # Application entry point
├── go.mod                 # Go dependencies
└── wails.json             # Wails project configuration
```

### Application icon

The canonical application icon is:

```text
ui/icon.svg
```

The build process rasterizes this SVG into Wails' required `build/appicon.png` asset. Wails then generates the Windows icon resources used by the executable.

Do not replace the SVG with a generated `.ico` file. The SVG is the source asset; generated build files are disposable.

## Development

Windows is required for building and running the desktop application.

Install:

- Go 1.26+
- Node.js 22+
- ImageMagick
- Wails v2.15.0

Install the pinned Wails CLI:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
```

Check the environment:

```powershell
wails doctor
```

Run the application in development mode:

```powershell
wails dev
```

Build a production executable:

```powershell
wails build -clean
```

The executable is written to:

```text
build/bin/mk-link.exe
```

## Link type details

### Symbolic link

A symbolic link is a filesystem reference to another file or directory.

Use it when:

- The link should behave as a flexible reference.
- You are working with development or compatibility paths.
- The target may be a file or directory.

### Hard link

A hard link is another directory entry for the same file data.

Restrictions:

- Files only.
- The target and link must be on the same volume.

### Junction

A junction is a Windows directory reparse point.

Use it when:

- You need to redirect a local directory.
- You want applications to continue using the original path.

Restrictions:

- Directories only.
- Local targets only; network/UNC targets are rejected.

## Release builds

Pushing a version tag such as `v2.0.1` starts the Windows release workflow.

The workflow:

1. Installs the pinned Wails CLI.
2. Runs `wails doctor`.
3. Installs and builds the frontend.
4. Downloads Go dependencies.
5. Runs `go test ./...` and `go vet ./...`.
6. Builds `mk-link.exe` with the canonical SVG icon.
7. Verifies that the executable exists.
8. Publishes `mk-link.exe` to the GitHub release.

The release workflow is intentionally Windows-only because MK-Link is a Windows desktop application.

## Manual release checks

Before publishing a release, verify the following on a real Windows 10/11 machine:

- `go test ./...`
- `go vet ./...`
- `wails doctor`
- `wails build -clean`
- File symbolic-link creation
- Directory symbolic-link creation
- Hard-link creation and same-volume rejection
- Junction creation and UNC rejection
- Existing destination rejection
- Invalid and reserved names
- Unicode and long paths
- UAC/access-denied behavior
- Drag and drop from Explorer
- Native picker behavior
- Open-in-Explorer behavior

## License

See [License](License).
