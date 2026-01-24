package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"mklink/styles"
	"mklink/utils"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FilePickerMode determines what files to show
type FilePickerMode int

const (
	FilePickerFile FilePickerMode = iota
	FilePickerDir
	FilePickerAny // File or Dir
)

// FileEntry represents a file or directory
type FileEntry struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}

// FilePickerModel handles file selection
type FilePickerModel struct {
	mode         FilePickerMode
	currentDir   string
	entries      []FileEntry
	cursor       int
	selectedPath string
	done         bool
	cancelled    bool
	err          error

	// For manual path input
	showInput bool
	pathInput textinput.Model
}

// NewFilePickerModel creates a new file picker
func NewFilePickerModel() *FilePickerModel {

	ti := textinput.New()
	ti.Placeholder = "Enter path..."
	ti.CharLimit = 500
	ti.Width = 60

	wd, _ := os.Getwd()

	fp := &FilePickerModel{
		mode:       FilePickerAny,
		currentDir: wd,
		pathInput:  ti,
	}
	fp.loadFiles()
	return fp
}

func (fp *FilePickerModel) SetMode(mode FilePickerMode) {
	fp.mode = mode
	fp.loadFiles()
}

func (fp *FilePickerModel) loadFiles() {
	fp.entries = []FileEntry{}
	fp.cursor = 0

	entries, err := os.ReadDir(fp.currentDir)
	if err != nil {
		fp.err = err
		return
	}

	var files []FileEntry

	// Parent directory option (if not root)
	// We handle ".." via Left Arrow, but showing it is nice visual context
	// keeping it simple for now, standard navigation is better.

	for _, entry := range entries {
		// Filter hidden files and system directories
		if strings.HasPrefix(entry.Name(), ".") || strings.HasPrefix(entry.Name(), "$") {
			continue
		}

		fullPath := filepath.Join(fp.currentDir, entry.Name())
		if isHidden(fullPath) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fe := FileEntry{
			Name:  entry.Name(),
			Path:  fullPath,
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		}
		files = append(files, fe)
	}

	// Sort: Directories first, then files
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir && !files[j].IsDir {
			return true
		}
		if !files[i].IsDir && files[j].IsDir {
			return false
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	fp.entries = files
}

func isHidden(path string) bool {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false
	}
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}

func looksLikePath(s string) bool {
	return (len(s) >= 3 && s[1] == ':' && (s[2] == '\\' || s[2] == '/')) ||
		strings.HasPrefix(s, "\\\\") ||
		(strings.Contains(s, "\\") && strings.Contains(s, "."))
}

func (fp *FilePickerModel) Update(msg tea.Msg) (*FilePickerModel, tea.Cmd) {
	var cmd tea.Cmd

	// Input Mode
	if fp.showInput {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				path := strings.Trim(fp.pathInput.Value(), "'\"")
				if path == "" {
					return fp, nil
				}

				fi, err := os.Stat(path)
				if err != nil {
					fp.err = fmt.Errorf("not found: %s", path)
					return fp, nil
				}

				// Validation
				valid := false
				if fp.mode == FilePickerAny {
					valid = true
				}
				if fp.mode == FilePickerDir && fi.IsDir() {
					valid = true
				}
				if fp.mode == FilePickerFile && !fi.IsDir() {
					valid = true
				}

				if valid {
					fp.selectedPath = path
					fp.done = true
				} else {
					fp.err = fmt.Errorf("invalid type for current mode")
				}
				return fp, nil

			case "esc":
				fp.showInput = false
				fp.pathInput.SetValue("")
				fp.err = nil
				return fp, nil
			}
		}
		fp.pathInput, cmd = fp.pathInput.Update(msg)
		return fp, cmd
	}

	// Browser Mode
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()

		if looksLikePath(keyStr) {
			fp.showInput = true
			fp.err = nil
			fp.pathInput.Focus()
			fp.pathInput.SetValue(keyStr)
			fp.pathInput.CursorEnd()
			return fp, textinput.Blink
		}

		switch keyStr {
		case "up", "k":
			if fp.cursor > 0 {
				fp.cursor--
			}
		case "down", "j":
			if fp.cursor < len(fp.entries)-1 {
				fp.cursor++
			}

		case "right", "l":
			// Enter directory (Navigation)
			if len(fp.entries) > 0 {
				selected := fp.entries[fp.cursor]
				if selected.IsDir {
					fp.currentDir = selected.Path
					fp.loadFiles()
				}
			}

		case "left", "h":
			// Go up
			parent := filepath.Dir(fp.currentDir)
			if parent != fp.currentDir {
				fp.currentDir = parent
				fp.loadFiles()
			}

		case "enter":
			// Select Target
			if len(fp.entries) > 0 {
				selected := fp.entries[fp.cursor]

				// Validation Logic
				// If mode is Dir, we can select a Dir.
				// If mode is File, we can select a File.
				// If mode is Any, any.

				valid := false
				if fp.mode == FilePickerAny {
					valid = true
				}
				if fp.mode == FilePickerDir && selected.IsDir {
					valid = true
				}
				if fp.mode == FilePickerFile && !selected.IsDir {
					valid = true
				}

				if valid {
					fp.selectedPath = selected.Path
					fp.done = true
				} else {
					// If we try to select a Dir in File mode, maybe enter it?
					if selected.IsDir {
						fp.currentDir = selected.Path
						fp.loadFiles()
					}
				}
			}

		case "p":
			fp.showInput = true
			fp.err = nil
			fp.pathInput.Focus()
			return fp, textinput.Blink

		case "esc":
			fp.cancelled = true
			fp.done = true
		}
	}
	return fp, nil
}

func (fp *FilePickerModel) View() string {
	var b strings.Builder

	// Header
	icon := styles.IconFolder
	title := "Select Target"
	if fp.mode == FilePickerDir {
		title = "Select Directory"
	}
	if fp.mode == FilePickerFile {
		title = "Select File"
	}

	b.WriteString(styles.Header.Render(fmt.Sprintf(" %s %s ", icon, title)))
	b.WriteString("\n\n")

	b.WriteString(styles.InputLabel.Render("Path: ") + fp.currentDir + "\n\n")

	if fp.err != nil {
		b.WriteString(styles.Error.Render("Error: "+fp.err.Error()) + "\n\n")
	}

	if fp.showInput {
		b.WriteString(styles.InputLabel.Render("Manual Input: "))
		b.WriteString(fp.pathInput.View() + "\n\n")
		b.WriteString(styles.Help.Render("Enter path • Enter to confirm • Esc cancel"))
		return b.String()
	}

	// List
	if len(fp.entries) == 0 {
		b.WriteString(styles.Warning.Render("  (Empty Directory)") + "\n\n")
	} else {
		// Windowing
		limit := 10
		start := 0
		if fp.cursor >= limit {
			start = fp.cursor - limit + 1
		}
		end := start + limit
		if end > len(fp.entries) {
			end = len(fp.entries)
		}

		for i := start; i < end; i++ {
			entry := fp.entries[i]
			cursor := "  "
			style := styles.FileItem

			if i == fp.cursor {
				cursor = styles.IconPointer + " "
				style = styles.SelectedFile
			}

			icon := styles.IconFile
			if entry.IsDir {
				icon = styles.IconFolder
				if i != fp.cursor {
					style = styles.DirItem.Copy().PaddingLeft(4)
				}
			}

			label := fmt.Sprintf("%s %s", icon, entry.Name)
			if !entry.IsDir {
				label += fmt.Sprintf(" (%s)", utils.FormatSize(entry.Size))
			}

			b.WriteString(style.Render(cursor+label) + "\n")
		}

		if len(fp.entries) > limit {
			b.WriteString(styles.Help.Render(fmt.Sprintf("\n  ... %d more items", len(fp.entries)-end)))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(styles.Help.Render("↑↓ Nav • → Enter Dir • ← Up Dir • Enter Select • p Path"))

	return b.String()
}
