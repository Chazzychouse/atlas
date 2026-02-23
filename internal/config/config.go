package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	IMAPHost string
	IMAPPort int
	IMAPUser string
	IMAPPass string

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string

	FromName  string
	FromEmail string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		IMAPHost: os.Getenv("ATLAS_IMAP_HOST"),
		IMAPPort: 993,
		IMAPUser: os.Getenv("ATLAS_IMAP_USER"),
		IMAPPass: os.Getenv("ATLAS_IMAP_PASS"),

		SMTPHost: os.Getenv("ATLAS_SMTP_HOST"),
		SMTPPort: 587,
		SMTPUser: os.Getenv("ATLAS_SMTP_USER"),
		SMTPPass: os.Getenv("ATLAS_SMTP_PASS"),

		FromName:  os.Getenv("ATLAS_FROM_NAME"),
		FromEmail: os.Getenv("ATLAS_FROM_EMAIL"),
	}

	if v := os.Getenv("ATLAS_IMAP_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid ATLAS_IMAP_PORT: %w", err)
		}
		cfg.IMAPPort = port
	}

	if v := os.Getenv("ATLAS_SMTP_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid ATLAS_SMTP_PORT: %w", err)
		}
		cfg.SMTPPort = port
	}

	if cfg.IMAPHost == "" {
		return nil, fmt.Errorf("ATLAS_IMAP_HOST is required")
	}
	if cfg.IMAPUser == "" {
		return nil, fmt.Errorf("ATLAS_IMAP_USER is required")
	}
	if cfg.IMAPPass == "" {
		return nil, fmt.Errorf("ATLAS_IMAP_PASS is required")
	}
	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("ATLAS_SMTP_HOST is required")
	}
	if cfg.SMTPUser == "" {
		return nil, fmt.Errorf("ATLAS_SMTP_USER is required")
	}
	if cfg.SMTPPass == "" {
		return nil, fmt.Errorf("ATLAS_SMTP_PASS is required")
	}
	if cfg.FromEmail == "" {
		return nil, fmt.Errorf("ATLAS_FROM_EMAIL is required")
	}

	return cfg, nil
}
