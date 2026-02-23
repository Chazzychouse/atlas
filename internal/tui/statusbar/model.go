package statusbar

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the status bar with spinner and text.
type Model struct {
	spinner  spinner.Model
	text     string
	isError  bool
	spinning bool
	width    int
}

// New creates a new status bar.
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	return Model{
		spinner: s,
		text:    "Ready",
	}
}

// SetWidth sets the status bar width.
func (m *Model) SetWidth(w int) {
	m.width = w
}

// SetStatus updates the status text.
func (m *Model) SetStatus(text string, isError bool) {
	m.text = text
	m.isError = isError
}

// StartSpinner begins the loading spinner.
func (m *Model) StartSpinner() tea.Cmd {
	m.spinning = true
	return m.spinner.Tick
}

// StopSpinner stops the loading spinner.
func (m *Model) StopSpinner() {
	m.spinning = false
}

// Update handles spinner ticks.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.spinning {
		return m, nil
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// View renders the status bar.
func (m Model) View() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Width(m.width)

	if m.isError {
		style = style.Foreground(lipgloss.Color("#FF4444"))
	}

	text := m.text
	if m.spinning {
		text = m.spinner.View() + " " + text
	}

	return style.Render(text)
}
