package reader

import (
	"fmt"
	"strings"

	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// Model is the message reader view.
type Model struct {
	imap     *mail.IMAPClient
	viewport viewport.Model
	message  *mail.Message
	folder   string
	uid      uint32
	width    int
	height   int
	keys     KeyMap
	loading  bool
	ready    bool
}

// New creates a new reader view.
func New(imap *mail.IMAPClient, folder string, uid uint32, width, height int) *Model {
	return &Model{
		imap:   imap,
		folder: folder,
		uid:    uid,
		width:  width,
		height: height,
		keys:   DefaultKeyMap(),
	}
}

func (m *Model) Init() tea.Cmd {
	m.loading = true
	return tea.Batch(
		fetchMessage(m.imap, m.folder, m.uid),
		func() tea.Msg { return tui.SpinnerStartMsg{} },
		func() tea.Msg { return tui.StatusMsg{Text: "Loading message..."} },
	)
}

func (m *Model) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.ready {
			m.viewport.Width = m.width
			m.viewport.Height = m.height - 2
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Reply):
			if m.message != nil {
				return m, func() tea.Msg {
					return tui.PushViewMsg{
						ViewID:  tui.ViewComposer,
						ReplyTo: m.message,
					}
				}
			}
		case key.Matches(msg, m.keys.ReplyAll):
			if m.message != nil {
				return m, func() tea.Msg {
					return tui.PushViewMsg{
						ViewID:   tui.ViewComposer,
						ReplyTo:  m.message,
						ReplyAll: true,
					}
				}
			}
		case key.Matches(msg, m.keys.Forward):
			if m.message != nil {
				return m, func() tea.Msg {
					return tui.PushViewMsg{
						ViewID:  tui.ViewComposer,
						ReplyTo: m.message,
						Forward: true,
					}
				}
			}
		case key.Matches(msg, m.keys.Delete):
			return m, tea.Batch(
				deleteMessage(m.imap, m.uid),
				func() tea.Msg { return tui.StatusMsg{Text: "Deleting..."} },
			)
		}

	case tui.MessageLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			return m, tea.Batch(
				func() tea.Msg { return tui.SpinnerStopMsg{} },
				func() tea.Msg {
					return tui.StatusMsg{Text: "Error: " + msg.Err.Error(), IsError: true}
				},
			)
		}
		m.message = msg.Message
		m.setupViewport()
		return m, tea.Batch(
			func() tea.Msg { return tui.SpinnerStopMsg{} },
			func() tea.Msg { return tui.StatusMsg{Text: m.message.Subject} },
		)

	case tui.MessageDeletedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Delete failed: " + msg.Err.Error(), IsError: true}
			}
		}
		return m, func() tea.Msg { return tui.PopViewMsg{} }
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	if m.loading {
		return "\n  Loading message..."
	}
	if !m.ready {
		return ""
	}
	return m.viewport.View()
}

func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{
		m.keys.Reply, m.keys.ReplyAll, m.keys.Forward,
		m.keys.Delete,
	}
}

func (m *Model) setupViewport() {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	var sb strings.Builder
	sb.WriteString(headerStyle.Render(m.message.Subject))
	sb.WriteString("\n")
	sb.WriteString(labelStyle.Render("From: ") + m.message.From + "\n")
	sb.WriteString(labelStyle.Render("To:   ") + strings.Join(m.message.To, ", ") + "\n")
	if len(m.message.Cc) > 0 {
		sb.WriteString(labelStyle.Render("Cc:   ") + strings.Join(m.message.Cc, ", ") + "\n")
	}
	sb.WriteString(labelStyle.Render("Date: ") + m.message.Date.Format("Mon, 02 Jan 2006 15:04") + "\n")
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("─", min(m.width, 80))))
	sb.WriteString("\n")

	// Word wrap the body
	wrapWidth := min(m.width-2, 80)
	if wrapWidth < 20 {
		wrapWidth = 20
	}
	wrapped := wordwrap.String(m.message.Body, wrapWidth)
	sb.WriteString(wrapped)

	m.viewport = viewport.New(m.width, m.height-2)
	m.viewport.SetContent(sb.String())
	m.ready = true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
