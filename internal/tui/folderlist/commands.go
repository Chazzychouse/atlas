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

func createFolder(imap *mail.IMAPClient, name string) tea.Cmd {
	return func() tea.Msg {
		err := imap.CreateFolder(name)
		return tui.FolderCreatedMsg{Name: name, Err: err}
	}
}

func deleteFolder(imap *mail.IMAPClient, name string) tea.Cmd {
	return func() tea.Msg {
		err := imap.DeleteFolder(name)
		return tui.FolderDeletedMsg{Name: name, Err: err}
	}
}

func renameFolder(imap *mail.IMAPClient, oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		err := imap.RenameFolder(oldName, newName)
		return tui.FolderRenamedMsg{OldName: oldName, NewName: newName, Err: err}
	}
}
