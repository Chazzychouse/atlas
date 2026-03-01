package mail

import (
	"strings"
	"testing"

	"github.com/chazzychouse/atlas/internal/config"
)

func TestExtractEmail(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user@example.com", "user@example.com"},
		{"Alice <alice@example.com>", "alice@example.com"},
		{"  user@example.com  ", "user@example.com"},
		{"Name <addr@host.com>", "addr@host.com"},
		{"<bare@host.com>", "bare@host.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractEmail(tt.input)
			if got != tt.want {
				t.Errorf("extractEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildMessage(t *testing.T) {
	cfg := &config.Config{
		FromName:  "Test",
		FromEmail: "test@example.com",
	}

	msg := &SendMessage{
		To:      []string{"alice@example.com"},
		Cc:      []string{"bob@example.com"},
		Subject: "Hello World",
		Body:    "Test body content",
	}

	raw := string(buildMessage(msg, cfg))

	for _, want := range []string{"From:", "To:", "Subject:", "Mime-Version:", "Content-Type:", "Cc:"} {
		if !strings.Contains(raw, want) {
			t.Errorf("buildMessage missing header %q", want)
		}
	}

	if !strings.Contains(raw, "Test body content") {
		t.Error("buildMessage missing body")
	}

	if !strings.Contains(raw, "alice@example.com") {
		t.Error("buildMessage missing To address")
	}
}

func TestBuildMessageUsesFromField(t *testing.T) {
	cfg := &config.Config{
		FromName:  "Config Name",
		FromEmail: "config@example.com",
	}

	msg := &SendMessage{
		From: "Override <override@example.com>",
		To:   []string{"alice@example.com"},
		Body: "body",
	}

	raw := string(buildMessage(msg, cfg))

	if !strings.Contains(raw, "override@example.com") {
		t.Error("buildMessage should use From field from SendMessage when set")
	}
}

func TestBuildMessageNoCc(t *testing.T) {
	cfg := &config.Config{
		FromEmail: "test@example.com",
	}

	msg := &SendMessage{
		To:   []string{"alice@example.com"},
		Body: "body",
	}

	raw := string(buildMessage(msg, cfg))

	if strings.Contains(raw, "Cc:") {
		t.Error("buildMessage should not include Cc header when no Cc addresses")
	}
}
