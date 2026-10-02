# 🔗 MK-Link Go

A modern, interactive Windows tool for creating symbolic links, hard links, and directory junctions using a terminal user interface (TUI).

## ✨ Features

- **📁 Interactive File/Directory Picker**: Browse and select target files or directories with an intuitive TUI interface
- **🔗 Multiple Link Types**:
  - **🔄 Symbolic Links**: Standard symlinks for files and directories
  - **🔗 Hard Links**: Direct file system links (files only)
  - **🔗 Junctions**: Directory links using Windows reparse points (directories only)
- **🔐 Automatic Admin Elevation**: Automatically requests administrative privileges when needed
- **✅ Path Validation**: Ensures target paths exist and validates link creation
- **💻 Modern TUI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a smooth, responsive experience

## 📋 Requirements

- 🪟 Windows 10/11 (or Windows Server with symbolic link support)
- 🐹 Go 1.19+ (for building from source)
- 🔑 Administrative privileges (required for creating links)

## 📥 Installation

### 📦 Option 1: Download Pre-built Binary

📥 Download the latest `mklink.exe` from the [releases page](https://github.com/nameIess/MK-link/releases).

### 🛠️ Option 2: Build from Source

1. Clone the repository:

   ```bash
   git clone https://github.com/nameIess/MK-link.git
   cd MK-link
   ```

2. Build the executable:

   ### Option A: Using the build script (recommended)

   ```bash
   .\build.bat
   ```

   This will automatically handle icon embedding, optimization, and UPX packing.

   ### Option B: Manual build

   ```bash
   # Generate resource file from icon.ico (if available)
   rsrc -ico ui/icon.ico -o mklink.syso

   # Build optimized executable
   go build -ldflags="-s -w" -o mklink.exe

   # Pack with UPX (optional, helps with antivirus)
   upx --best mklink.exe
   ```

## 🚀 Usage

1. Run the executable:

   ```bash
   .\mklink.exe
   ```

2. If not running as administrator, the tool will automatically request elevation

3. Follow the interactive prompts:
   - **Select Target**: Browse and choose the file or directory to link to
   - **Choose Link Type**: Select from available options based on target type
   - **Enter Link Name**: Provide a name for the new link
   - **Choose Location**: Select where to create the link
   - **Confirm**: Review and confirm the link creation

## 🔗 Link Types Explained

### For Files:

- **File Symlink**: Creates a symbolic link that points to the original file
- **Hard Link**: Creates a direct link to the file's data on disk

### For Directories:

- **Directory Symlink**: Creates a symbolic link that points to the original directory
- **Junction**: Creates a directory junction using Windows reparse points (local volumes only)

## 🏗️ Building

To build the project:

```bash
go build -ldflags="-s -w" -o mklink.exe
```

The `mklink.syso` file contains the application icon and should be in the same directory as `main.go`.

## 📦 Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - Bubble Tea components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly (especially on different Windows versions)
5. Submit a pull request

## 📜 License

MIT License - see [LICENSE](License) file for details.

## ❓ Troubleshooting

- **"Access Denied"**: Ensure you're running as administrator
- **"Symbolic links not supported"**: Enable developer mode in Windows settings or use junctions for directories
- **UI Issues**: Ensure your terminal supports ANSI colors and has a compatible font
- **Antivirus Flagging**: Go binaries may trigger false positives due to their nature. The build process strips debug info to minimize this. If issues persist, add the executable to your antivirus exclusions or consider code signing.

## 🙏 Credits

Built with [Charm](https://charm.sh/) libraries for the terminal interface.</content>
<parameter name="filePath">d:\gemini-workspace\MK-Link\README.md
