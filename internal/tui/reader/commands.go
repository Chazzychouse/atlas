package reader

import (
	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func fetchMessage(imap *mail.IMAPClient, folder string, uid uint32) tea.Cmd {
	return func() tea.Msg {
		msg, err := imap.FetchMessage(folder, uid)
		return tui.MessageLoadedMsg{Message: msg, Err: err}
	}
}

func deleteMessage(imap *mail.IMAPClient, uid uint32) tea.Cmd {
	return func() tea.Msg {
		err := imap.DeleteMessage(uid)
		return tui.MessageDeletedMsg{UID: uid, Err: err}
	}
}
