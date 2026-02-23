package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/textproto"
	"strings"
	"time"

	"github.com/chazzychouse/atlas/internal/config"
	"github.com/emersion/go-sasl"

	gosmtp "net/smtp"
)

// SMTPClient sends email via SMTP with STARTTLS.
type SMTPClient struct {
	cfg *config.Config
}

// NewSMTPClient creates a new SMTP sender.
func NewSMTPClient(cfg *config.Config) *SMTPClient {
	return &SMTPClient{cfg: cfg}
}

// Send sends an email message.
func (s *SMTPClient) Send(msg *SendMessage) error {
	addr := net.JoinHostPort(s.cfg.SMTPHost, fmt.Sprintf("%d", s.cfg.SMTPPort))

	// Connect
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("SMTP connect: %w", err)
	}

	host := s.cfg.SMTPHost
	c, err := gosmtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("SMTP client: %w", err)
	}
	defer c.Close()

	// STARTTLS
	tlsConfig := &tls.Config{ServerName: host}
	if err := c.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("SMTP STARTTLS: %w", err)
	}

	// Auth using PLAIN via go-sasl
	_ = sasl.NewPlainClient("", s.cfg.SMTPUser, s.cfg.SMTPPass)
	auth := plainAuth(s.cfg.SMTPUser, s.cfg.SMTPPass, host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth: %w", err)
	}

	// Set sender
	from := msg.From
	if from == "" {
		from = s.cfg.FromEmail
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}

	// Set recipients
	allRecipients := append(msg.To, msg.Cc...)
	for _, rcpt := range allRecipients {
		addr := extractEmail(rcpt)
		if err := c.Rcpt(addr); err != nil {
			return fmt.Errorf("SMTP RCPT TO %s: %w", addr, err)
		}
	}

	// Build and write message
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}

	raw := buildMessage(msg, s.cfg)
	if _, err := w.Write(raw); err != nil {
		return fmt.Errorf("writing message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("closing DATA: %w", err)
	}

	return c.Quit()
}

func buildMessage(msg *SendMessage, cfg *config.Config) []byte {
	var buf bytes.Buffer

	from := msg.From
	if from == "" {
		from = FormatAddress(cfg.FromName, cfg.FromEmail)
	}

	h := textproto.MIMEHeader{}
	h.Set("From", from)
	h.Set("To", strings.Join(msg.To, ", "))
	if len(msg.Cc) > 0 {
		h.Set("Cc", strings.Join(msg.Cc, ", "))
	}
	h.Set("Subject", mime.QEncoding.Encode("utf-8", msg.Subject))
	h.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700"))
	h.Set("MIME-Version", "1.0")
	h.Set("Content-Type", "text/plain; charset=UTF-8")

	for key, vals := range h {
		for _, v := range vals {
			fmt.Fprintf(&buf, "%s: %s\r\n", key, v)
		}
	}
	buf.WriteString("\r\n")
	buf.WriteString(msg.Body)

	return buf.Bytes()
}

// extractEmail pulls the email address from "Name <email>" or bare "email" format.
func extractEmail(s string) string {
	if idx := strings.Index(s, "<"); idx >= 0 {
		end := strings.Index(s, ">")
		if end > idx {
			return s[idx+1 : end]
		}
	}
	return strings.TrimSpace(s)
}

// plainAuth implements smtp.Auth for PLAIN authentication.
type smtpPlainAuth struct {
	username, password, host string
}

func plainAuth(username, password, host string) gosmtp.Auth {
	return &smtpPlainAuth{username, password, host}
}

func (a *smtpPlainAuth) Start(server *gosmtp.ServerInfo) (string, []byte, error) {
	resp := []byte("\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *smtpPlainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, fmt.Errorf("unexpected server challenge")
	}
	return nil, nil
}
