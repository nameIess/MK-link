package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	PrimaryColor   = lipgloss.Color("#7C3AED") // Purple
	SecondaryColor = lipgloss.Color("#10B981") // Green
	AccentColor    = lipgloss.Color("#F59E0B") // Amber
	ErrorColor     = lipgloss.Color("#EF4444") // Red
	SubtleColor    = lipgloss.Color("#6B7280") // Gray
	TextColor      = lipgloss.Color("#F3F4F6") // Light gray
)

// Styles
var (
	// Title styles
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginBottom(1)

	// Header box style
	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(TextColor).
		Background(PrimaryColor).
		Padding(0, 2). // Standard padding
		MarginBottom(1)

	// Menu item styles
	MenuItem = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItem = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(PrimaryColor).
			Bold(true)

	// Box style for sections
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2)

	// DocStyle for global layout
	DocStyle = lipgloss.NewStyle().
			Padding(1, 2) // Simple margin

	// --- Missing Styles ---
	Description = lipgloss.NewStyle().
			Foreground(SubtleColor).
			Italic(true).
			MarginTop(1)

	Success = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true)

	Error = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	Warning = lipgloss.NewStyle().
		Foreground(AccentColor)

	InputLabel = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	Help = lipgloss.NewStyle().
		Foreground(SubtleColor).
		MarginTop(1)

	Progress = lipgloss.NewStyle().
			Foreground(SecondaryColor)

	FileItem = lipgloss.NewStyle().
			PaddingLeft(4)

	SelectedFile = lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(SecondaryColor).
			Bold(true)

	DirItem = lipgloss.NewStyle().
		Foreground(AccentColor)
)

// Icons
const (
	IconFolder   = "📁"
	IconFile     = "📄"
	IconLink     = "🔗"
	IconJunction = "⛩️"
	IconCheck    = "✓"
	IconCross    = "✗"
	IconArrow    = "→"
	IconPointer  = "▶"
	IconSettings = "⚙️"
	IconExit     = "❌"
	IconSuccess  = "✅"
	IconError    = "❌"
	IconWarning  = "⚠️"
)
