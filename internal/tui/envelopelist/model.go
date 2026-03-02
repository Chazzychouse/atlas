package envelopelist

import (
	"fmt"
	"strings"

	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the envelope list view.
type Model struct {
	imap      *mail.IMAPClient
	table     table.Model
	envelopes []mail.Envelope
	folder    string
	page      int
	total     uint32
	width     int
	height    int
	keys      KeyMap
	loading   bool
	selected  map[uint32]bool // UID -> selected
}

// New creates a new envelope list view.
func New(imap *mail.IMAPClient, folder string, width, height int) *Model {
	if folder == "" {
		folder = "INBOX"
	}

	columns := []table.Column{
		{Title: " ", Width: 3},
		{Title: "From", Width: 25},
		{Title: "Subject", Width: 40},
		{Title: "Date", Width: 18},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(height-2),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#4A4A4A")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(false)
	t.SetStyles(s)

	return &Model{
		imap:     imap,
		table:    t,
		folder:   folder,
		keys:     DefaultKeyMap(),
		width:    width,
		height:   height,
		selected: make(map[uint32]bool),
	}
}

func (m *Model) Init() tea.Cmd {
	m.loading = true
	return tea.Batch(
		fetchEnvelopes(m.imap, m.folder, m.page),
		func() tea.Msg {
			return tui.SpinnerStartMsg{}
		},
		func() tea.Msg {
			return tui.StatusMsg{Text: fmt.Sprintf("Loading %s...", m.folder)}
		},
	)
}

func (m *Model) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Select):
			if env := m.selectedEnvelope(); env != nil {
				if m.selected[env.UID] {
					delete(m.selected, env.UID)
				} else {
					m.selected[env.UID] = true
				}
				m.updateRows()
				// Move cursor down
				m.table, _ = m.table.Update(tea.KeyMsg{Type: tea.KeyDown})
			}
			return m, nil
		case key.Matches(msg, m.keys.SelectAll):
			if len(m.selected) > 0 {
				m.selected = make(map[uint32]bool)
			} else {
				for _, env := range m.envelopes {
					m.selected[env.UID] = true
				}
			}
			m.updateRows()
			return m, nil
		case key.Matches(msg, m.keys.Open):
			if env := m.selectedEnvelope(); env != nil {
				return m, func() tea.Msg {
					return tui.PushViewMsg{
						ViewID:      tui.ViewReader,
						EnvelopeUID: env.UID,
						Folder:      m.folder,
					}
				}
			}
		case key.Matches(msg, m.keys.Compose):
			return m, func() tea.Msg {
				return tui.PushViewMsg{ViewID: tui.ViewComposer}
			}
		case key.Matches(msg, m.keys.Delete):
			if uids := m.selectedUIDs(); len(uids) > 0 {
				m.selected = make(map[uint32]bool)
				return m, bulkDelete(m.imap, uids)
			}
			if env := m.selectedEnvelope(); env != nil {
				return m, deleteMessage(m.imap, env.UID)
			}
		case key.Matches(msg, m.keys.ToggleRead):
			if uids := m.selectedUIDs(); len(uids) > 0 {
				seen := m.bulkSeenTarget()
				m.selected = make(map[uint32]bool)
				return m, bulkSetSeen(m.imap, uids, seen)
			}
			if env := m.selectedEnvelope(); env != nil {
				return m, toggleRead(m.imap, env.UID, env.Seen)
			}
		case key.Matches(msg, m.keys.Reply):
			if env := m.selectedEnvelope(); env != nil {
				return m, func() tea.Msg {
					return tui.PushViewMsg{
						ViewID:      tui.ViewReader,
						EnvelopeUID: env.UID,
						Folder:      m.folder,
					}
				}
			}
		case key.Matches(msg, m.keys.NextPage):
			maxPage := int(m.total) / mail.PageSize
			if m.page < maxPage {
				m.page++
				m.loading = true
				m.selected = make(map[uint32]bool)
				return m, fetchEnvelopes(m.imap, m.folder, m.page)
			}
		case key.Matches(msg, m.keys.PrevPage):
			if m.page > 0 {
				m.page--
				m.loading = true
				m.selected = make(map[uint32]bool)
				return m, fetchEnvelopes(m.imap, m.folder, m.page)
			}
		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			m.selected = make(map[uint32]bool)
			return m, fetchEnvelopes(m.imap, m.folder, m.page)
		}

	case tui.EnvelopesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Error: " + msg.Err.Error(), IsError: true}
			}
		}
		m.envelopes = msg.Envelopes
		m.total = msg.Total
		m.updateRows()
		return m, tea.Batch(
			func() tea.Msg { return tui.SpinnerStopMsg{} },
			func() tea.Msg {
				return tui.StatusMsg{Text: fmt.Sprintf("%s (%d messages)", m.folder, m.total)}
			},
		)

	case tui.MessageDeletedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Delete failed: " + msg.Err.Error(), IsError: true}
			}
		}
		// Refresh the list
		return m, tea.Batch(
			fetchEnvelopes(m.imap, m.folder, m.page),
			func() tea.Msg {
				return tui.StatusMsg{Text: "Message deleted"}
			},
		)

	case tui.BulkDeletedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Bulk delete failed: " + msg.Err.Error(), IsError: true}
			}
		}
		return m, tea.Batch(
			fetchEnvelopes(m.imap, m.folder, m.page),
			func() tea.Msg {
				return tui.StatusMsg{Text: fmt.Sprintf("%d messages deleted", len(msg.UIDs))}
			},
		)

	case tui.BulkFlagUpdatedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Bulk flag update failed: " + msg.Err.Error(), IsError: true}
			}
		}
		uidSet := make(map[uint32]bool, len(msg.UIDs))
		for _, uid := range msg.UIDs {
			uidSet[uid] = true
		}
		for i := range m.envelopes {
			if uidSet[m.envelopes[i].UID] {
				m.envelopes[i].Seen = msg.Seen
			}
		}
		m.updateRows()
		label := "read"
		if !msg.Seen {
			label = "unread"
		}
		return m, func() tea.Msg {
			return tui.StatusMsg{Text: fmt.Sprintf("%d messages marked %s", len(msg.UIDs), label)}
		}

	case tui.FlagUpdatedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Flag update failed: " + msg.Err.Error(), IsError: true}
			}
		}
		// Update local state
		for i := range m.envelopes {
			if m.envelopes[i].UID == msg.UID {
				m.envelopes[i].Seen = msg.Seen
			}
		}
		m.updateRows()
		return m, nil
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.loading && len(m.envelopes) == 0 {
		return "\n  Loading..."
	}
	if !m.loading && len(m.envelopes) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(2, 4)
		return emptyStyle.Render("No messages in " + m.folder + "\n\nPress 'c' to compose a new message.")
	}
	return m.table.View()
}

func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{
		m.keys.Open, m.keys.Compose, m.keys.Delete,
		m.keys.ToggleRead, m.keys.Reply, m.keys.Select, m.keys.SelectAll,
		m.keys.NextPage, m.keys.PrevPage, m.keys.Refresh,
	}
}

func (m *Model) selectedEnvelope() *mail.Envelope {
	row := m.table.Cursor()
	if row >= 0 && row < len(m.envelopes) {
		return &m.envelopes[row]
	}
	return nil
}

// selectedUIDs returns UIDs of all selected messages, or nil if none selected.
func (m *Model) selectedUIDs() []uint32 {
	if len(m.selected) == 0 {
		return nil
	}
	uids := make([]uint32, 0, len(m.selected))
	for uid := range m.selected {
		uids = append(uids, uid)
	}
	return uids
}

// bulkSeenTarget returns the target seen state for bulk toggle:
// if any selected message is unread, mark all as read; otherwise mark all as unread.
func (m *Model) bulkSeenTarget() bool {
	for _, env := range m.envelopes {
		if m.selected[env.UID] && !env.Seen {
			return true // at least one unread → mark all as read
		}
	}
	return false // all are read → mark all as unread
}

func (m *Model) updateTableSize() {
	columns := []table.Column{
		{Title: " ", Width: 3},
		{Title: "From", Width: max(15, m.width/4)},
		{Title: "Subject", Width: max(20, m.width/2-10)},
		{Title: "Date", Width: 18},
	}
	m.table.SetColumns(columns)
	m.table.SetHeight(m.height - 2)
}

func (m *Model) updateRows() {
	rows := make([]table.Row, len(m.envelopes))
	for i, env := range m.envelopes {
		sel := " "
		if m.selected[env.UID] {
			sel = ">"
		}
		unread := " "
		if !env.Seen {
			unread = "*"
		}
		flag := sel + unread
		from := truncate(env.From, max(15, m.width/4))
		subj := truncate(env.Subject, max(20, m.width/2-10))
		date := env.Date.Format("Jan 02 15:04")
		rows[i] = table.Row{flag, from, subj, date}
	}
	m.table.SetRows(rows)
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
