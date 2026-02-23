package folderlist

import (
	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	return &Model{
		imap:   imap,
		list:   l,
		width:  width,
		height: height,
		keys:   DefaultKeyMap(),
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
		if key.Matches(msg, m.keys.Select) {
			if item, ok := m.list.SelectedItem().(folderItem); ok {
				return m, func() tea.Msg {
					return tui.FolderSelectedMsg{Folder: item.name}
				}
			}
		}

	case tui.FoldersLoadedMsg:
		if msg.Err != nil {
			return m, func() tea.Msg {
				return tui.StatusMsg{Text: "Folder load error: " + msg.Err.Error(), IsError: true}
			}
		}
		m.folders = msg.Folders
		items := make([]list.Item, len(m.folders))
		for i, f := range m.folders {
			items[i] = folderItem{name: f.Name}
		}
		m.list.SetItems(items)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return m.list.View()
}

func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{m.keys.Select}
}
