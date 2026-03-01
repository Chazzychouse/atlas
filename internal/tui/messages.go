package tui

import "github.com/chazzychouse/atlas/internal/mail"

// Navigation messages

// PushViewMsg requests navigating to a new view.
type PushViewMsg struct {
	ViewID ViewID
	// Optional data passed to the view being pushed.
	EnvelopeUID uint32      // for reader: which message to open
	ReplyTo     *mail.Message // for composer: reply/forward prefill
	ReplyAll    bool
	Forward     bool
	Folder      string // for folder selection
}

// PopViewMsg requests going back to the previous view.
type PopViewMsg struct{}

// Data messages

// FoldersLoadedMsg carries the list of folders from IMAP.
type FoldersLoadedMsg struct {
	Folders []mail.Folder
	Err     error
}

// EnvelopesLoadedMsg carries fetched envelopes.
type EnvelopesLoadedMsg struct {
	Envelopes []mail.Envelope
	Total     uint32
	Err       error
}

// MessageLoadedMsg carries a fully fetched message.
type MessageLoadedMsg struct {
	Message *mail.Message
	Err     error
}

// MessageSentMsg indicates the result of sending an email.
type MessageSentMsg struct {
	Err error
}

// MessageDeletedMsg indicates a message was deleted.
type MessageDeletedMsg struct {
	UID uint32
	Err error
}

// BulkDeletedMsg indicates multiple messages were deleted.
type BulkDeletedMsg struct {
	UIDs []uint32
	Err  error
}

// BulkFlagUpdatedMsg indicates flags were updated on multiple messages.
type BulkFlagUpdatedMsg struct {
	UIDs []uint32
	Seen bool
	Err  error
}

// MessageMovedMsg indicates a message was moved.
type MessageMovedMsg struct {
	UID    uint32
	Folder string
	Err    error
}

// FlagUpdatedMsg indicates a flag was updated.
type FlagUpdatedMsg struct {
	UID  uint32
	Seen bool
	Err  error
}

// FolderSelectedMsg indicates the user selected a folder.
type FolderSelectedMsg struct {
	Folder string
}

// Status messages

// StatusMsg updates the status bar text.
type StatusMsg struct {
	Text    string
	IsError bool
}

// SpinnerTickMsg is an alias for the spinner's internal tick.
type SpinnerStartMsg struct{}
type SpinnerStopMsg struct{}

// StatusClearMsg clears the status bar if the generation matches.
type StatusClearMsg struct{ Gen uint64 }

// FolderCreatedMsg indicates a folder was created.
type FolderCreatedMsg struct {
	Name string
	Err  error
}

// FolderDeletedMsg indicates a folder was deleted.
type FolderDeletedMsg struct {
	Name string
	Err  error
}

// FolderRenamedMsg indicates a folder was renamed.
type FolderRenamedMsg struct {
	OldName string
	NewName string
	Err     error
}

// ErrMsg is a generic error message.
type ErrMsg struct {
	Err error
}

// ContactsSyncedMsg carries envelopes fetched purely for contact extraction.
type ContactsSyncedMsg struct {
	Envelopes []mail.Envelope
}
