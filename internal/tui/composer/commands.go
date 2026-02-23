package composer

import (
	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func sendMessage(smtp *mail.SMTPClient, msg *mail.SendMessage) tea.Cmd {
	return func() tea.Msg {
		err := smtp.Send(msg)
		return tui.MessageSentMsg{Err: err}
	}
}
