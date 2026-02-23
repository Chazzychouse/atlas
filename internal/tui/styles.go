package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#6C6C6C")
	ColorAccent    = lipgloss.Color("#FF6F61")
	ColorMuted     = lipgloss.Color("#4A4A4A")
	ColorBright    = lipgloss.Color("#FFFFFF")
	ColorDim       = lipgloss.Color("#888888")
	ColorError     = lipgloss.Color("#FF4444")
	ColorSuccess   = lipgloss.Color("#44FF44")
)

// Shared styles
var (
	StyleStatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(ColorBright).
			Padding(0, 1)

	StyleStatusError = lipgloss.NewStyle().
				Background(lipgloss.Color("#333333")).
				Foreground(ColorError).
				Padding(0, 1)

	StyleSidebar = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderForeground(ColorMuted)

	StyleTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleUnread = lipgloss.NewStyle().
			Bold(true)

	StyleDim = lipgloss.NewStyle().
			Foreground(ColorDim)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorSecondary)
)
