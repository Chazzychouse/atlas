package mail

import "time"

// Envelope represents the summary of an email message.
type Envelope struct {
	UID     uint32
	From    string
	To      []string
	Cc      []string
	Subject string
	Date    time.Time
	Seen    bool
}

// Message represents a full email message with body content.
type Message struct {
	Envelope
	Body string // plain text body
}

// Folder represents an IMAP mailbox folder.
type Folder struct {
	Name       string
	Attributes []string
}

// SendMessage holds the data for composing and sending an email.
type SendMessage struct {
	From    string
	To      []string
	Cc      []string
	Subject string
	Body    string
}
