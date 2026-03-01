package cmd

import (
	"fmt"
	"strings"

	"github.com/chazzychouse/atlas/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard",
	Long:  "Configure Atlas by walking through each setting interactively.",
	RunE:  runSetup,
}

func runSetup(_ *cobra.Command, _ []string) error {
	path, err := config.Path()
	if err != nil {
		return err
	}

	m := newSetupModel()
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return err
	}

	sm := result.(setupModel)
	if sm.cancelled {
		fmt.Println("Setup cancelled.")
		return nil
	}

	cfg := sm.toConfig()
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Config saved to %s\n", path)
	return nil
}

type field struct {
	label    string
	input    textinput.Model
	password bool
}

type setupModel struct {
	fields    []field
	cursor    int
	submitted bool
	cancelled bool
}

func newSetupModel() setupModel {
	newInput := func(placeholder string, password bool) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.CharLimit = 256
		if password {
			ti.EchoMode = textinput.EchoPassword
		}
		return ti
	}

	fields := []field{
		{label: "IMAP Host", input: newInput("imap.gmail.com", false)},
		{label: "IMAP Port", input: newInput("993", false)},
		{label: "IMAP User", input: newInput("you@example.com", false)},
		{label: "IMAP Password", input: newInput("", true), password: true},
		{label: "SMTP Host", input: newInput("smtp.gmail.com", false)},
		{label: "SMTP Port", input: newInput("587", false)},
		{label: "SMTP User", input: newInput("you@example.com", false)},
		{label: "SMTP Password", input: newInput("", true), password: true},
		{label: "From Name", input: newInput("Your Name", false)},
		{label: "From Email", input: newInput("you@example.com", false)},
	}

	fields[0].input.Focus()

	return setupModel{fields: fields}
}

func (m setupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyEnter, tea.KeyTab, tea.KeyDown:
			if m.cursor < len(m.fields)-1 {
				m.fields[m.cursor].input.Blur()
				m.cursor++
				m.fields[m.cursor].input.Focus()
				return m, textinput.Blink
			}
			// Last field — submit
			m.submitted = true
			return m, tea.Quit

		case tea.KeyShiftTab, tea.KeyUp:
			if m.cursor > 0 {
				m.fields[m.cursor].input.Blur()
				m.cursor--
				m.fields[m.cursor].input.Focus()
				return m, textinput.Blink
			}
		}
	}

	var cmd tea.Cmd
	m.fields[m.cursor].input, cmd = m.fields[m.cursor].input.Update(msg)
	return m, cmd
}

var (
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(16)
	activeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true).Width(16)
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
)

func (m setupModel) View() string {
	var sb strings.Builder
	sb.WriteString("\n  Atlas Setup\n")
	sb.WriteString(hintStyle.Render("  ↑/↓ or Tab to navigate · Enter on last field to save · Esc to cancel"))
	sb.WriteString("\n\n")

	for i, f := range m.fields {
		style := labelStyle
		if i == m.cursor {
			style = activeStyle
		}
		sb.WriteString("  " + style.Render(f.label+":") + " " + f.input.View() + "\n")
	}

	sb.WriteString("\n")
	if m.cursor == len(m.fields)-1 {
		sb.WriteString(hintStyle.Render("  Press Enter to save"))
	}
	sb.WriteString("\n")
	return sb.String()
}

func (m setupModel) toConfig() *config.Config {
	val := func(i int) string { return m.fields[i].input.Value() }
	portOrDefault := func(i, def int) int {
		var p int
		if _, err := fmt.Sscan(val(i), &p); err != nil || p == 0 {
			return def
		}
		return p
	}

	return &config.Config{
		IMAPHost:  val(0),
		IMAPPort:  portOrDefault(1, 993),
		IMAPUser:  val(2),
		IMAPPass:  val(3),
		SMTPHost:  val(4),
		SMTPPort:  portOrDefault(5, 587),
		SMTPUser:  val(6),
		SMTPPass:  val(7),
		FromName:  val(8),
		FromEmail: val(9),
	}
}
