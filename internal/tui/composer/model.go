package composer

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/chazzychouse/atlas/internal/config"
	"github.com/chazzychouse/atlas/internal/contacts"
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

// suggWindowSize is how many suggestions to show at once.
const suggWindowSize = 3

// suggBoxWidth is the inner content width of the suggestion dropdown.
const suggBoxWidth = 42

// Model is the email composer view.
type Model struct {
	cfg         *config.Config
	smtp        *mail.SMTPClient
	contacts    *contacts.Manager
	inputs      []textinput.Model
	body        textarea.Model
	focus       int
	width       int
	height      int
	keys        KeyMap
	sending     bool
	suggestions []contacts.Contact
	suggSel     int // absolute index into suggestions
	showSugg    bool
}

// New creates a new composer view.
func New(cfg *config.Config, smtp *mail.SMTPClient, ctcts *contacts.Manager, width, height int) *Model {
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
		cfg:      cfg,
		smtp:     smtp,
		contacts: ctcts,
		inputs:   inputs,
		body:     body,
		keys:     DefaultKeyMap(),
		width:    width,
		height:   height,
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

		// When suggestions are visible on a To/Cc field, intercept navigation keys.
		if m.showSugg && (m.focus == fieldTo || m.focus == fieldCc) {
			switch msg.String() {
			case "down":
				m.suggSel = (m.suggSel + 1) % len(m.suggestions)
				return m, nil
			case "up":
				m.suggSel = (m.suggSel - 1 + len(m.suggestions)) % len(m.suggestions)
				return m, nil
			case "tab", "enter":
				m.acceptSuggestion()
				return m, nil
			case "esc":
				m.showSugg = false
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, m.keys.Send):
			return m, m.send()
		case key.Matches(msg, m.keys.NextField):
			m.showSugg = false
			m.nextField()
			return m, nil
		case key.Matches(msg, m.keys.PrevField):
			m.showSugg = false
			m.prevField()
			return m, nil
		case key.Matches(msg, m.keys.Cancel):
			return m, func() tea.Msg { return tui.PopViewMsg{} }
		}

	case tui.MessageSentMsg:
		m.sending = false
		if msg.Err != nil {
			return m, tea.Batch(
				func() tea.Msg { return tui.SpinnerStopMsg{} },
				func() tea.Msg {
					return tui.StatusMsg{Text: "Send failed: " + msg.Err.Error(), IsError: true}
				},
			)
		}
		return m, tea.Batch(
			func() tea.Msg { return tui.SpinnerStopMsg{} },
			func() tea.Msg { return tui.StatusMsg{Text: "Message sent!"} },
			func() tea.Msg { return tui.PopViewMsg{} },
		)
	}

	cmd := m.updateInputs(msg)

	// After processing input, refresh suggestions if a To/Cc field is active.
	if m.focus == fieldTo || m.focus == fieldCc {
		m.updateSuggestions()
	}

	return m, cmd
}

func (m *Model) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	padStyle := lipgloss.NewStyle().PaddingLeft(2)

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("  Compose"))
	sb.WriteString("\n\n")

	// Render To, Cc, Subject individually so we can inject suggestion boxes.
	for i, input := range m.inputs {
		sb.WriteString("  " + input.View() + "\n")
		if m.showSugg && (i == fieldTo || i == fieldCc) && m.focus == i {
			sb.WriteString(m.renderSuggestions())
		}
	}

	sb.WriteString("\n")
	sb.WriteString(padStyle.Render(m.body.View()))

	sb.WriteString("\n\n")
	if m.showSugg {
		sb.WriteString(labelStyle.Render("  Ctrl+S: send │ ↑↓: cycle │ ↩/Tab: pick │ Esc: dismiss"))
	} else {
		sb.WriteString(labelStyle.Render("  Ctrl+S: send | Tab: next field | Esc: cancel"))
	}

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

	// Validate all email addresses
	for _, addr := range to {
		if !mail.ValidateEmail(addr) {
			return func() tea.Msg {
				return tui.StatusMsg{Text: fmt.Sprintf("Invalid To address: %s", addr), IsError: true}
			}
		}
	}
	cc := parseAddresses(m.inputs[fieldCc].Value())
	for _, addr := range cc {
		if !mail.ValidateEmail(addr) {
			return func() tea.Msg {
				return tui.StatusMsg{Text: fmt.Sprintf("Invalid Cc address: %s", addr), IsError: true}
			}
		}
	}

	m.sending = true
	sendMsg := &mail.SendMessage{
		From:    mail.FormatAddress(m.cfg.FromName, m.cfg.FromEmail),
		To:      to,
		Cc:      cc,
		Subject: m.inputs[fieldSubject].Value(),
		Body:    m.body.Value(),
	}

	// Record sent addresses in contacts.
	if m.contacts != nil {
		now := time.Now()
		for _, addr := range append(to, cc...) {
			m.contacts.Update(addr, now)
		}
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

// updateSuggestions recomputes the suggestion list from the active field's last token.
func (m *Model) updateSuggestions() {
	if m.contacts == nil {
		return
	}
	token := lastToken(m.inputs[m.focus].Value())
	results := m.contacts.Search(token)
	if len(results) == 0 {
		m.showSugg = false
		m.suggestions = nil
		return
	}
	// Keep selection stable when results are the same length, reset otherwise.
	if len(results) != len(m.suggestions) {
		m.suggSel = 0
	}
	m.suggestions = results
	m.showSugg = true
}

// acceptSuggestion replaces the last token in the focused field with the selected contact.
func (m *Model) acceptSuggestion() {
	if !m.showSugg || len(m.suggestions) == 0 {
		return
	}
	c := m.suggestions[m.suggSel]
	current := m.inputs[m.focus].Value()

	var prefix string
	if idx := strings.LastIndex(current, ","); idx >= 0 {
		prefix = strings.TrimRight(current[:idx+1], " ") + " "
	}

	m.inputs[m.focus].SetValue(prefix + c.Formatted() + ", ")
	// Move cursor to end.
	m.inputs[m.focus].CursorEnd()

	m.showSugg = false
	m.suggestions = nil
}

// suggestionWindowStart returns the scroll offset so suggSel is always visible.
func (m *Model) suggestionWindowStart() int {
	n := len(m.suggestions)
	if n <= suggWindowSize {
		return 0
	}
	start := m.suggSel - suggWindowSize + 1
	if start < 0 {
		start = 0
	}
	if start > n-suggWindowSize {
		start = n - suggWindowSize
	}
	return start
}

// renderSuggestions builds the inline dropdown string.
func (m *Model) renderSuggestions() string {
	if len(m.suggestions) == 0 {
		return ""
	}

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	moreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	start := m.suggestionWindowStart()
	end := start + suggWindowSize
	if end > len(m.suggestions) {
		end = len(m.suggestions)
	}

	innerWidth := suggBoxWidth

	var lines []string

	// "▲ N above" indicator
	if start > 0 {
		label := fmt.Sprintf("▲ %d above", start)
		lines = append(lines, moreStyle.Render(padOrTrunc(label, innerWidth)))
	}

	for i := start; i < end; i++ {
		c := m.suggestions[i]
		text := c.Formatted()
		text = padOrTrunc(text, innerWidth)
		if i == m.suggSel {
			lines = append(lines, selectedStyle.Render(text))
		} else {
			lines = append(lines, dimStyle.Render(text))
		}
	}

	// "▼ N more" indicator
	remaining := len(m.suggestions) - end
	if remaining > 0 {
		label := fmt.Sprintf("▼ %d more", remaining)
		lines = append(lines, moreStyle.Render(padOrTrunc(label, innerWidth)))
	}

	box := boxStyle.Render(strings.Join(lines, "\n"))

	// Indent to align with the text input area (2 prefix + 9 prompt = 11 chars).
	indent := strings.Repeat(" ", 11)
	var out strings.Builder
	for _, line := range strings.Split(box, "\n") {
		out.WriteString(indent + line + "\n")
	}
	return out.String()
}

// padOrTrunc pads or truncates s to exactly width visible runes.
func padOrTrunc(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n > width {
		runes := []rune(s)
		return string(runes[:width-1]) + "…"
	}
	return s + strings.Repeat(" ", width-n)
}

// lastToken returns the text after the last comma in s (trimmed).
// Returns "" if s ends with a comma (user just accepted an address).
func lastToken(s string) string {
	idx := strings.LastIndex(s, ",")
	if idx < 0 {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(s[idx+1:])
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
