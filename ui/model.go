package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"mklink/logic"
	"mklink/styles"
)

type Step int

const (
	StepTarget Step = iota
	StepType
	StepName
	StepLocation
	StepConfirm
	StepDone
)

type MenuItem struct {
	Title string
	Desc  string
	Icon  string
	Type  logic.LinkType
}

type Model struct {
	Step         Step
	TargetPicker *FilePickerModel
	LocPicker    *FilePickerModel

	// Layout
	Width  int
	Height int

	// Inputs
	NameInput textinput.Model

	// Menu State
	Cursor int

	// Data
	TargetPath string
	LinkType   logic.LinkType
	LinkName   string
	LinkLoc    string

	// State helpers
	targetIsDir bool
	err         error
	msg         string
}

func InitialModel() Model {
	// Name Input
	ni := textinput.New()
	ni.Placeholder = "Link Name"
	ni.Width = 30

	return Model{
		Step:         StepTarget,
		TargetPicker: NewFilePickerModel(),
		LocPicker:    NewFilePickerModel(),
		NameInput:    ni,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Recalculate layout if needed
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	// Route updates based on step
	switch m.Step {
	case StepTarget:
		var newPicker *FilePickerModel
		newPicker, cmd = m.TargetPicker.Update(msg)
		m.TargetPicker = newPicker

		if m.TargetPicker.done {
			if m.TargetPicker.cancelled {
				return m, tea.Quit
			}
			m.TargetPath = m.TargetPicker.selectedPath
			fi, _ := os.Stat(m.TargetPath)
			m.targetIsDir = fi.IsDir()

			m.NameInput.SetValue(filepath.Base(m.TargetPath))

			m.LocPicker.SetMode(FilePickerDir)
			wd, _ := os.Getwd()
			m.LocPicker.currentDir = wd
			m.LocPicker.loadFiles()

			m.Step++
		}
		return m, cmd

	case StepType:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "up", "k":
				m.Cursor--
				if m.Cursor < 0 {
					m.Cursor = len(m.getMenuItems()) - 1
				}
			case "down", "j":
				m.Cursor++
				if m.Cursor >= len(m.getMenuItems()) {
					m.Cursor = 0
				}
			case "enter":
				items := m.getMenuItems()
				m.LinkType = items[m.Cursor].Type
				m.Step++
				m.NameInput.Focus()
				return m, textinput.Blink
			case "esc":
				m.Step = StepTarget
				m.TargetPicker.done = false
			}
		}

	case StepName:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.Type {
			case tea.KeyEnter:
				val := strings.TrimSpace(m.NameInput.Value())
				if val != "" {
					m.LinkName = val
					m.Step++
				}
			case tea.KeyEsc:
				m.Step = StepType
			}
		}
		m.NameInput, cmd = m.NameInput.Update(msg)
		return m, cmd

	case StepLocation:
		var newPicker *FilePickerModel
		newPicker, cmd = m.LocPicker.Update(msg)
		m.LocPicker = newPicker

		if m.LocPicker.done {
			if m.LocPicker.cancelled {
				m.Step = StepName
				m.LocPicker.done = false
				m.LocPicker.cancelled = false
				return m, nil
			}
			m.LinkLoc = m.LocPicker.selectedPath
			m.Step++
		}
		return m, cmd

	case StepConfirm:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y", "Y", "enter":
				fullPath := filepath.Join(m.LinkLoc, m.LinkName)
				err := logic.CreateLink(m.LinkType, m.TargetPath, fullPath)
				if err != nil {
					m.err = err
					m.msg = err.Error()
				} else {
					m.msg = fmt.Sprintf("Link created at: %s", fullPath)
				}
				m.Step = StepDone
			case "n", "N", "esc":
				m.Step = StepLocation
				m.LocPicker.done = false
			}
		}

	case StepDone:
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.Type == tea.KeyEsc || msg.Type == tea.KeyEnter {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	// Render the current content
	content := m.renderContent()

	// Apply global padding/margins (DocStyle)
	return styles.DocStyle.Render(content)
}

func (m Model) renderContent() string {
	var b strings.Builder

	// Banner
	if m.Step != StepDone {
		b.WriteString(styles.Header.Render(" 🔗 MK-LINK GO "))
		b.WriteString("\n\n")
	}

	switch m.Step {
	case StepTarget:
		b.WriteString(m.TargetPicker.View())

	case StepType:
		b.WriteString(styles.InputLabel.Render("Select Link Type:"))
		b.WriteString("\n\n")

		items := m.getMenuItems()
		for i, item := range items {
			cursor := "  "
			style := styles.MenuItem
			if i == m.Cursor {
				cursor = styles.IconPointer + " "
				style = styles.SelectedItem
			}
			b.WriteString(style.Render(cursor+item.Icon+" "+item.Title) + "\n")

			if i == m.Cursor {
				b.WriteString(styles.Description.Render("    "+item.Desc) + "\n")
			}
			b.WriteString("\n")
		}
		b.WriteString(styles.Help.Render("↑↓ Navigate • Enter Select • Esc Back"))

	case StepName:
		b.WriteString(styles.Title.Render("🏷  Name Your Link"))
		b.WriteString("\n\n")
		b.WriteString(styles.InputLabel.Render("Name: "))
		b.WriteString(m.NameInput.View())
		b.WriteString("\n\n")
		b.WriteString(styles.Description.Render("Target: " + filepath.Base(m.TargetPath)))
		b.WriteString("\n\n")
		b.WriteString(styles.Help.Render("Enter Confirm • Esc Back"))

	case StepLocation:
		b.WriteString(styles.Title.Render("📍 Select Destination Folder"))
		b.WriteString("\n\n")
		b.WriteString(m.LocPicker.View())

	case StepConfirm:
		b.WriteString(styles.InputLabel.Render("Summary"))
		b.WriteString("\n\n")

		summary := fmt.Sprintf(
			"Target:   %s\nType:     %s\nName:     %s\nLocation: %s",
			m.TargetPath,
			m.getTypeName(),
			m.LinkName,
			m.LinkLoc,
		)
		b.WriteString(styles.Box.Render(summary))
		b.WriteString("\n\n")
		b.WriteString(styles.Warning.Render("Create this link? (Y/n)"))

	case StepDone:
		b.WriteString("\n")
		if m.err != nil {
			b.WriteString(styles.Error.Render(styles.IconError + " Failed"))
			b.WriteString("\n\n")
			b.WriteString(m.err.Error())
		} else {
			b.WriteString(styles.Success.Render(styles.IconSuccess + " Success!"))
			b.WriteString("\n\n")
			b.WriteString(m.msg)
		}
		b.WriteString("\n\n")
		b.WriteString(styles.Help.Render("Press Esc/Enter to quit"))
	}

	return b.String()
}

func (m Model) getMenuItems() []MenuItem {
	if m.targetIsDir {
		return []MenuItem{
			{"Directory Symlink", "Standard symbolic link", styles.IconFolder, logic.LinkTypeDirSymlink},
			{"Junction", "Harder link (Local only)", styles.IconJunction, logic.LinkTypeJunction},
		}
	}
	return []MenuItem{
		{"File Symlink", "Standard symbolic link", styles.IconFile, logic.LinkTypeFileSymlink},
		{"Hard Link", "Merged file entry", styles.IconLink, logic.LinkTypeHardLink},
	}
}

func (m Model) getTypeName() string {
	switch m.LinkType {
	case logic.LinkTypeDirSymlink:
		return "Dir Symlink"
	case logic.LinkTypeJunction:
		return "Junction"
	case logic.LinkTypeFileSymlink:
		return "File Symlink"
	case logic.LinkTypeHardLink:
		return "Hard Link"
	default:
		return "Unknown"
	}
}
