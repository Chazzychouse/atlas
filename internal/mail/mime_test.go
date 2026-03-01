package mail

import (
	"strings"
	"testing"
)

func TestParsePlainTextSimple(t *testing.T) {
	raw := "Content-Type: text/plain; charset=utf-8\r\n" +
		"Subject: Test\r\n" +
		"\r\n" +
		"Hello, this is plain text."

	body, err := ParsePlainText([]byte(raw))
	if err != nil {
		t.Fatalf("ParsePlainText() error: %v", err)
	}

	if !strings.Contains(body, "Hello, this is plain text.") {
		t.Errorf("ParsePlainText() = %q, want to contain plain text body", body)
	}
}

func TestParsePlainTextPrefersPlainOverHTML(t *testing.T) {
	raw := "MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"boundary\"\r\n" +
		"\r\n" +
		"--boundary\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		"Plain text version\r\n" +
		"--boundary\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		"<p>HTML version</p>\r\n" +
		"--boundary--\r\n"

	body, err := ParsePlainText([]byte(raw))
	if err != nil {
		t.Fatalf("ParsePlainText() error: %v", err)
	}

	if !strings.Contains(body, "Plain text version") {
		t.Errorf("ParsePlainText() should prefer text/plain, got %q", body)
	}
	if strings.Contains(body, "HTML version") {
		t.Error("ParsePlainText() should not return HTML when text/plain is available")
	}
}

func TestParsePlainTextFallsBackToHTML(t *testing.T) {
	raw := "MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"boundary\"\r\n" +
		"\r\n" +
		"--boundary\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		"<p>HTML only content</p>\r\n" +
		"--boundary--\r\n"

	body, err := ParsePlainText([]byte(raw))
	if err != nil {
		t.Fatalf("ParsePlainText() error: %v", err)
	}

	if body == "" {
		t.Error("ParsePlainText() should fall back to HTML when no text/plain")
	}
	// html-to-markdown should have converted it
	if strings.Contains(body, "<p>") {
		t.Errorf("ParsePlainText() should convert HTML to markdown, got %q", body)
	}
}

func TestParsePlainTextEmpty(t *testing.T) {
	raw := "Content-Type: multipart/alternative; boundary=\"boundary\"\r\n" +
		"\r\n" +
		"--boundary\r\n" +
		"Content-Type: application/octet-stream\r\n" +
		"\r\n" +
		"binary data\r\n" +
		"--boundary--\r\n"

	body, err := ParsePlainText([]byte(raw))
	if err != nil {
		t.Fatalf("ParsePlainText() error: %v", err)
	}

	if body != "" {
		t.Errorf("ParsePlainText() = %q, want empty for non-text parts", body)
	}
}

func TestFormatAddressFromMime(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"", "user@example.com", "user@example.com"},
		{"Alice", "alice@example.com", "Alice <alice@example.com>"},
		{"Bob Smith", "bob@test.com", "Bob Smith <bob@test.com>"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := FormatAddress(tt.name, tt.email)
			if got != tt.want {
				t.Errorf("FormatAddress(%q, %q) = %q, want %q", tt.name, tt.email, got, tt.want)
			}
		})
	}
}
