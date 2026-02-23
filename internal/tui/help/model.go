package help

import (
	"strings"

	"github.com/chazzychouse/atlas/internal/tui"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the help overlay view.
type Model struct {
	sections []tui.HelpSection
	width    int
	height   int
}

// New creates a new help view.
func New(sections []tui.HelpSection, width, height int) *Model {
	return &Model{sections: sections, width: width, height: height}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *Model) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Width(14)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(title.Render("  Key Bindings"))
	sb.WriteString("\n\n")

	for _, section := range m.sections {
		sb.WriteString(title.Render("  " + section.Title))
		sb.WriteString("\n")
		for _, b := range section.Bindings {
			if !b.Enabled() {
				continue
			}
			h := b.Help()
			if h.Key == "" {
				continue
			}
			sb.WriteString("  ")
			sb.WriteString(keyStyle.Render(h.Key))
			sb.WriteString(descStyle.Render(h.Desc))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "close help")),
	}
}
