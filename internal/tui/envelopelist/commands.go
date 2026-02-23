package envelopelist

import (
	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func fetchEnvelopes(imap *mail.IMAPClient, folder string, page int) tea.Cmd {
	return func() tea.Msg {
		total, err := imap.SelectFolder(folder)
		if err != nil {
			return tui.EnvelopesLoadedMsg{Err: err}
		}

		if total == 0 {
			return tui.EnvelopesLoadedMsg{Total: 0}
		}

		// Calculate range (newest first): fetch from end of mailbox
		pageSize := uint32(mail.PageSize)
		end := total - uint32(page)*pageSize
		start := uint32(1)
		if end > pageSize {
			start = end - pageSize + 1
		}
		if end == 0 {
			return tui.EnvelopesLoadedMsg{Total: total}
		}

		envelopes, err := imap.FetchEnvelopes(start, end)
		if err != nil {
			return tui.EnvelopesLoadedMsg{Err: err}
		}

		// Reverse to show newest first
		for i, j := 0, len(envelopes)-1; i < j; i, j = i+1, j-1 {
			envelopes[i], envelopes[j] = envelopes[j], envelopes[i]
		}

		return tui.EnvelopesLoadedMsg{
			Envelopes: envelopes,
			Total:     total,
		}
	}
}

func deleteMessage(imap *mail.IMAPClient, uid uint32) tea.Cmd {
	return func() tea.Msg {
		err := imap.DeleteMessage(uid)
		return tui.MessageDeletedMsg{UID: uid, Err: err}
	}
}

func toggleRead(imap *mail.IMAPClient, uid uint32, currentlySeen bool) tea.Cmd {
	return func() tea.Msg {
		newSeen := !currentlySeen
		err := imap.SetSeen(uid, newSeen)
		return tui.FlagUpdatedMsg{UID: uid, Seen: newSeen, Err: err}
	}
}

func bulkDelete(imap *mail.IMAPClient, uids []uint32) tea.Cmd {
	return func() tea.Msg {
		err := imap.DeleteMessages(uids)
		return tui.BulkDeletedMsg{UIDs: uids, Err: err}
	}
}

func bulkSetSeen(imap *mail.IMAPClient, uids []uint32, seen bool) tea.Cmd {
	return func() tea.Msg {
		err := imap.SetSeenBulk(uids, seen)
		return tui.BulkFlagUpdatedMsg{UIDs: uids, Seen: seen, Err: err}
	}
}
