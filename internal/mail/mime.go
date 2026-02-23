package mail

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/emersion/go-message/mail"
)

// ParsePlainText extracts the plain text body from a raw RFC5322 message.
func ParsePlainText(raw []byte) (string, error) {
	mr, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("creating mail reader: %w", err)
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("reading mail part: %w", err)
		}

		switch p.Header.(type) {
		case *mail.InlineHeader:
			ct, _, _ := p.Header.(*mail.InlineHeader).ContentType()
			if strings.HasPrefix(ct, "text/plain") {
				body, err := io.ReadAll(p.Body)
				if err != nil {
					return "", fmt.Errorf("reading body: %w", err)
				}
				return string(body), nil
			}
		}
	}

	return "", nil
}

// FormatAddress formats a name and email into "Name <email>" form.
func FormatAddress(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
