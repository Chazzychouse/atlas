package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "atlas", "config.toml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		t.Fatal(err)
	}

	content := `
imap_host = "imap.example.com"
imap_port = 993
imap_user = "user@example.com"
imap_pass = "secret"
smtp_host = "smtp.example.com"
smtp_port = 587
smtp_user = "user@example.com"
smtp_pass = "secret"
from_name = "Test User"
from_email = "user@example.com"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// Override XDG_CONFIG_HOME so Path() returns our temp dir
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.IMAPHost != "imap.example.com" {
		t.Errorf("IMAPHost = %q, want %q", cfg.IMAPHost, "imap.example.com")
	}
	if cfg.IMAPPort != 993 {
		t.Errorf("IMAPPort = %d, want 993", cfg.IMAPPort)
	}
	if cfg.SMTPPort != 587 {
		t.Errorf("SMTPPort = %d, want 587", cfg.SMTPPort)
	}
	if cfg.FromName != "Test User" {
		t.Errorf("FromName = %q, want %q", cfg.FromName, "Test User")
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "atlas", "config.toml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		t.Fatal(err)
	}

	content := `
imap_host = "imap.example.com"
imap_user = "user@example.com"
imap_pass = "secret"
smtp_host = "smtp.example.com"
smtp_user = "user@example.com"
smtp_pass = "secret"
from_email = "user@example.com"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("ATLAS_IMAP_HOST", "override.example.com")
	t.Setenv("ATLAS_SMTP_PORT", "465")
	t.Setenv("ATLAS_FROM_NAME", "Env User")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.IMAPHost != "override.example.com" {
		t.Errorf("IMAPHost = %q, want env override %q", cfg.IMAPHost, "override.example.com")
	}
	if cfg.SMTPPort != 465 {
		t.Errorf("SMTPPort = %d, want env override 465", cfg.SMTPPort)
	}
	if cfg.FromName != "Env User" {
		t.Errorf("FromName = %q, want env override %q", cfg.FromName, "Env User")
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should fail when config file is missing")
	}
}

func TestLoadInvalidSMTPPort(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "atlas", "config.toml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		t.Fatal(err)
	}

	content := `
imap_host = "imap.example.com"
imap_user = "user@example.com"
imap_pass = "secret"
smtp_host = "smtp.example.com"
smtp_user = "user@example.com"
smtp_pass = "secret"
from_email = "user@example.com"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("ATLAS_SMTP_PORT", "notanumber")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should fail with invalid ATLAS_SMTP_PORT")
	}
}

func TestValidateMissingFields(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{"missing imap_host", Config{IMAPUser: "u", IMAPPass: "p", SMTPHost: "s", SMTPUser: "u", SMTPPass: "p", FromEmail: "e"}},
		{"missing imap_user", Config{IMAPHost: "h", IMAPPass: "p", SMTPHost: "s", SMTPUser: "u", SMTPPass: "p", FromEmail: "e"}},
		{"missing imap_pass", Config{IMAPHost: "h", IMAPUser: "u", SMTPHost: "s", SMTPUser: "u", SMTPPass: "p", FromEmail: "e"}},
		{"missing smtp_host", Config{IMAPHost: "h", IMAPUser: "u", IMAPPass: "p", SMTPUser: "u", SMTPPass: "p", FromEmail: "e"}},
		{"missing smtp_user", Config{IMAPHost: "h", IMAPUser: "u", IMAPPass: "p", SMTPHost: "s", SMTPPass: "p", FromEmail: "e"}},
		{"missing smtp_pass", Config{IMAPHost: "h", IMAPUser: "u", IMAPPass: "p", SMTPHost: "s", SMTPUser: "u", FromEmail: "e"}},
		{"missing from_email", Config{IMAPHost: "h", IMAPUser: "u", IMAPPass: "p", SMTPHost: "s", SMTPUser: "u", SMTPPass: "p"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validate(&tt.cfg); err == nil {
				t.Error("validate() should fail")
			}
		})
	}
}

func TestValidateAllPresent(t *testing.T) {
	cfg := &Config{
		IMAPHost:  "h",
		IMAPUser:  "u",
		IMAPPass:  "p",
		SMTPHost:  "s",
		SMTPUser:  "u",
		SMTPPass:  "p",
		FromEmail: "e",
	}
	if err := validate(cfg); err != nil {
		t.Errorf("validate() unexpected error: %v", err)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &Config{
		IMAPHost:  "imap.test.com",
		IMAPPort:  993,
		IMAPUser:  "user@test.com",
		IMAPPass:  "pass",
		SMTPHost:  "smtp.test.com",
		SMTPPort:  587,
		SMTPUser:  "user@test.com",
		SMTPPass:  "pass",
		FromName:  "Tester",
		FromEmail: "user@test.com",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error after Save: %v", err)
	}

	if loaded.IMAPHost != cfg.IMAPHost {
		t.Errorf("IMAPHost = %q, want %q", loaded.IMAPHost, cfg.IMAPHost)
	}
	if loaded.FromName != cfg.FromName {
		t.Errorf("FromName = %q, want %q", loaded.FromName, cfg.FromName)
	}
}

func TestDefaultPorts(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "atlas", "config.toml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		t.Fatal(err)
	}

	// Config without port settings — should use defaults
	content := `
imap_host = "imap.example.com"
imap_user = "user@example.com"
imap_pass = "secret"
smtp_host = "smtp.example.com"
smtp_user = "user@example.com"
smtp_pass = "secret"
from_email = "user@example.com"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.IMAPPort != 993 {
		t.Errorf("default IMAPPort = %d, want 993", cfg.IMAPPort)
	}
	if cfg.SMTPPort != 587 {
		t.Errorf("default SMTPPort = %d, want 587", cfg.SMTPPort)
	}
}
