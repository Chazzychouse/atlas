package folderlist

import (
	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func fetchFolders(imap *mail.IMAPClient) tea.Cmd {
	return func() tea.Msg {
		folders, err := imap.ListFolders()
		return tui.FoldersLoadedMsg{Folders: folders, Err: err}
	}
}
