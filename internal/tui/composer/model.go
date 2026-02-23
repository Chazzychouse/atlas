package composer

import (
	"fmt"
	"strings"

	"github.com/chazzychouse/atlas/internal/config"
	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	fieldTo = iota
	fieldCc
	fieldSubject
	fieldBody
	fieldCount
)

// Model is the email composer view.
type Model struct {
	cfg      *config.Config
	smtp     *mail.SMTPClient
	inputs   []textinput.Model
	body     textarea.Model
	focus    int
	width    int
	height   int
	keys     KeyMap
	sending  bool
}

// New creates a new composer view.
func New(cfg *config.Config, smtp *mail.SMTPClient, width, height int) *Model {
	inputs := make([]textinput.Model, 3) // To, Cc, Subject

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].CharLimit = 500
	}
	inputs[fieldTo].Placeholder = "recipient@example.com"
	inputs[fieldTo].Prompt = "To:      "
	inputs[fieldCc].Placeholder = "cc@example.com"
	inputs[fieldCc].Prompt = "Cc:      "
	inputs[fieldSubject].Placeholder = "Subject"
	inputs[fieldSubject].Prompt = "Subject: "

	inputs[fieldTo].Focus()

	body := textarea.New()
	body.Placeholder = "Compose your message..."
	body.SetWidth(width - 2)
	body.SetHeight(height - 10)
	body.CharLimit = 0

	return &Model{
		cfg:    cfg,
		smtp:   smtp,
		inputs: inputs,
		body:   body,
		keys:   DefaultKeyMap(),
		width:  width,
		height: height,
	}
}

// Prefill sets up the composer for reply/forward.
func (m *Model) Prefill(msg *mail.Message, replyAll, forward bool) {
	if forward {
		m.inputs[fieldSubject].SetValue("Fwd: " + msg.Subject)
		m.body.SetValue(
			fmt.Sprintf("\n\n---------- Forwarded message ----------\nFrom: %s\nDate: %s\nSubject: %s\n\n%s",
				msg.From, msg.Date.Format("Mon, 02 Jan 2006 15:04"), msg.Subject, msg.Body),
		)
	} else {
		// Reply
		m.inputs[fieldTo].SetValue(msg.From)
		if replyAll && len(msg.To) > 0 {
			m.inputs[fieldCc].SetValue(strings.Join(msg.To, ", "))
		}
		subj := msg.Subject
		if !strings.HasPrefix(strings.ToLower(subj), "re:") {
			subj = "Re: " + subj
		}
		m.inputs[fieldSubject].SetValue(subj)

		// Quote original
		var quoted strings.Builder
		quoted.WriteString("\n\n")
		quoted.WriteString(fmt.Sprintf("On %s, %s wrote:\n",
			msg.Date.Format("Mon, 02 Jan 2006 15:04"), msg.From))
		for _, line := range strings.Split(msg.Body, "\n") {
			quoted.WriteString("> " + line + "\n")
		}
		m.body.SetValue(quoted.String())
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.body.SetWidth(m.width - 2)
		m.body.SetHeight(m.height - 10)

	case tea.KeyMsg:
		if m.sending {
			return m, nil
		}
		switch {
		case key.Matches(msg, m.keys.Send):
			return m, m.send()
		case key.Matches(msg, m.keys.NextField):
			m.nextField()
			return m, nil
		case key.Matches(msg, m.keys.PrevField):
			m.prevField()
			return m, nil
		case key.Matches(msg, m.keys.Cancel):
			return m, func() tea.Msg { return tui.PopViewMsg{} }
		}

	case tui.MessageSentMsg:
		m.sending = false
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Send failed: " + msg.Err.Error(), IsError: true}
			}
		}
		return m, tea.Batch(
			func() tea.Msg { return tui.StatusMsg{Text: "Message sent!"} },
			func() tea.Msg { return tui.PopViewMsg{} },
		)
	}

	return m, m.updateInputs(msg)
}

func (m *Model) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("  Compose"))
	sb.WriteString("\n\n")

	for i, input := range m.inputs {
		if i == m.focus {
			sb.WriteString("  " + input.View() + "\n")
		} else {
			sb.WriteString("  " + input.View() + "\n")
		}
	}

	sb.WriteString("\n")
	if m.focus == fieldBody {
		sb.WriteString("  " + m.body.View())
	} else {
		sb.WriteString("  " + m.body.View())
	}

	sb.WriteString("\n\n")
	sb.WriteString(labelStyle.Render("  Ctrl+S: send | Tab: next field | Esc: cancel"))

	if m.sending {
		sb.WriteString("\n\n  Sending...")
	}

	return sb.String()
}

func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{
		m.keys.Send, m.keys.NextField, m.keys.PrevField, m.keys.Cancel,
	}
}

func (m *Model) send() tea.Cmd {
	to := parseAddresses(m.inputs[fieldTo].Value())
	if len(to) == 0 {
		return func() tea.Msg {
			return tui.StatusMsg{Text: "To field is required", IsError: true}
		}
	}

	m.sending = true
	sendMsg := &mail.SendMessage{
		From:    mail.FormatAddress(m.cfg.FromName, m.cfg.FromEmail),
		To:      to,
		Cc:      parseAddresses(m.inputs[fieldCc].Value()),
		Subject: m.inputs[fieldSubject].Value(),
		Body:    m.body.Value(),
	}

	return tea.Batch(
		sendMessage(m.smtp, sendMsg),
		func() tea.Msg { return tui.SpinnerStartMsg{} },
		func() tea.Msg { return tui.StatusMsg{Text: "Sending..."} },
	)
}

func (m *Model) nextField() {
	m.blurAll()
	m.focus = (m.focus + 1) % fieldCount
	m.focusCurrent()
}

func (m *Model) prevField() {
	m.blurAll()
	m.focus = (m.focus - 1 + fieldCount) % fieldCount
	m.focusCurrent()
}

func (m *Model) blurAll() {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	m.body.Blur()
}

func (m *Model) focusCurrent() {
	if m.focus < len(m.inputs) {
		m.inputs[m.focus].Focus()
	} else {
		m.body.Focus()
	}
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if m.focus < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.body, cmd = m.body.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func parseAddresses(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var addrs []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			addrs = append(addrs, p)
		}
	}
	return addrs
}
