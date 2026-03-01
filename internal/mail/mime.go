package mail

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/emersion/go-message/mail"
)

// ParsePlainText extracts the plain text body from a raw RFC5322 message.
// If no text/plain part exists, it falls back to converting text/html to markdown.
func ParsePlainText(raw []byte) (string, error) {
	mr, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("creating mail reader: %w", err)
	}

	var htmlBody string

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
			body, err := io.ReadAll(p.Body)
			if err != nil {
				return "", fmt.Errorf("reading body: %w", err)
			}
			if strings.HasPrefix(ct, "text/plain") {
				return string(body), nil
			}
			if strings.HasPrefix(ct, "text/html") && htmlBody == "" {
				htmlBody = string(body)
			}
		}
	}

	if htmlBody != "" {
		md, err := htmltomarkdown.ConvertString(htmlBody)
		if err != nil {
			return "", fmt.Errorf("converting html to markdown: %w", err)
		}
		return md, nil
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
