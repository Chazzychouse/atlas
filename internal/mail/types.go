package mail

import (
	"net/mail"
	"strings"
	"time"
)

// ValidateEmail checks if s is a valid email address.
// It accepts both bare addresses and "Name <addr>" format.
func ValidateEmail(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	_, err := mail.ParseAddress(s)
	return err == nil
}

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
