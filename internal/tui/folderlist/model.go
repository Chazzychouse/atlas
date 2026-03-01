package folderlist

import (
	"strings"

	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	modeNormal mode = iota
	modeNew
	modeRename
	modeConfirmDelete
)

type folderItem struct {
	name string
}

func (f folderItem) Title() string       { return f.name }
func (f folderItem) Description() string { return "" }
func (f folderItem) FilterValue() string { return f.name }

// Model is the folder list sidebar view.
type Model struct {
	imap    *mail.IMAPClient
	list    list.Model
	folders []mail.Folder
	width   int
	height  int
	keys    KeyMap
	mode    mode
	input   textinput.Model
}

// New creates a new folder list view.
func New(imap *mail.IMAPClient, width, height int) *Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))

	l := list.New(nil, delegate, width, height)
	l.Title = "Folders"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	ti := textinput.New()
	ti.CharLimit = 128

	return &Model{
		imap:   imap,
		list:   l,
		width:  width,
		height: height,
		keys:   DefaultKeyMap(),
		input:  ti,
	}
}

func (m *Model) Init() tea.Cmd {
	return fetchFolders(m.imap)
}

func (m *Model) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(m.width, m.height)

	case tea.KeyMsg:
		if m.mode != modeNormal {
			return m.handleInputKey(msg)
		}

		switch {
		case key.Matches(msg, m.keys.Select):
			if item, ok := m.list.SelectedItem().(folderItem); ok {
				return m, func() tea.Msg {
					return tui.FolderSelectedMsg{Folder: item.name}
				}
			}

		case key.Matches(msg, m.keys.New):
			m.mode = modeNew
			m.input.Placeholder = "New folder name"
			m.input.SetValue("")
			m.input.Focus()
			return m, textinput.Blink

		case key.Matches(msg, m.keys.Rename):
			if item, ok := m.list.SelectedItem().(folderItem); ok {
				m.mode = modeRename
				m.input.Placeholder = item.name
				m.input.SetValue(item.name)
				m.input.Focus()
				return m, textinput.Blink
			}

		case key.Matches(msg, m.keys.Delete):
			if _, ok := m.list.SelectedItem().(folderItem); ok {
				m.mode = modeConfirmDelete
				m.input.Placeholder = "y/N"
				m.input.SetValue("")
				m.input.Focus()
				return m, textinput.Blink
			}
		}

	case tui.FoldersLoadedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Folder load error: " + msg.Err.Error(), IsError: true}
			}
		}
		m.folders = msg.Folders
		m.list.SetItems(m.toListItems())
		return m, nil

	case tui.FolderCreatedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Create failed: " + msg.Err.Error(), IsError: true}
			}
		}
		return m, tea.Batch(
			fetchFolders(m.imap),
			func() tea.Msg { return tui.StatusMsg{Text: "Created " + msg.Name} },
		)

	case tui.FolderDeletedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Delete failed: " + msg.Err.Error(), IsError: true}
			}
		}
		return m, tea.Batch(
			fetchFolders(m.imap),
			func() tea.Msg { return tui.StatusMsg{Text: "Deleted " + msg.Name} },
		)

	case tui.FolderRenamedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Rename failed: " + msg.Err.Error(), IsError: true}
			}
		}
		return m, tea.Batch(
			fetchFolders(m.imap),
			func() tea.Msg { return tui.StatusMsg{Text: "Renamed to " + msg.NewName} },
		)
	}

	var cmd tea.Cmd
	if m.mode == modeNormal {
		m.list, cmd = m.list.Update(msg)
	} else {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m *Model) handleInputKey(msg tea.KeyMsg) (tui.View, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelInput()
		return m, func() tea.Msg { return tui.StatusMsg{Text: ""} }

	case tea.KeyEnter:
		val := strings.TrimSpace(m.input.Value())
		currentMode := m.mode
		m.cancelInput()

		switch currentMode {
		case modeNew:
			if val == "" {
				return m, nil
			}
			return m, createFolder(m.imap, val)

		case modeRename:
			if val == "" {
				return m, nil
			}
			if item, ok := m.list.SelectedItem().(folderItem); ok {
				return m, renameFolder(m.imap, item.name, val)
			}

		case modeConfirmDelete:
			if strings.ToLower(val) == "y" {
				if item, ok := m.list.SelectedItem().(folderItem); ok {
					return m, deleteFolder(m.imap, item.name)
				}
			}
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) cancelInput() {
	m.mode = modeNormal
	m.input.Blur()
	m.input.SetValue("")
}

func (m *Model) View() string {
	if m.mode == modeNormal {
		return m.list.View()
	}

	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true).Padding(0, 1)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	var label string
	switch m.mode {
	case modeNew:
		label = "New folder:"
	case modeRename:
		label = "Rename to:"
	case modeConfirmDelete:
		if item, ok := m.list.SelectedItem().(folderItem); ok {
			label = "Delete " + item.name + "? (y/N):"
		}
	}

	var sb strings.Builder
	sb.WriteString(m.list.View())
	sb.WriteString("\n")
	sb.WriteString(promptStyle.Render(label) + " " + m.input.View())
	sb.WriteString("\n")
	sb.WriteString(hintStyle.Render("  Enter to confirm · Esc to cancel"))
	return sb.String()
}

func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{m.keys.Select, m.keys.New, m.keys.Rename, m.keys.Delete}
}

func (m *Model) toListItems() []list.Item {
	items := make([]list.Item, len(m.folders))
	for i, f := range m.folders {
		items[i] = folderItem{name: f.Name}
	}
	return items
}
