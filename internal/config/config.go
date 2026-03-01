package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all application configuration.
type Config struct {
	IMAPHost string `toml:"imap_host"`
	IMAPPort int    `toml:"imap_port"`
	IMAPUser string `toml:"imap_user"`
	IMAPPass string `toml:"imap_pass"`

	SMTPHost string `toml:"smtp_host"`
	SMTPPort int    `toml:"smtp_port"`
	SMTPUser string `toml:"smtp_user"`
	SMTPPass string `toml:"smtp_pass"`

	FromName  string `toml:"from_name"`
	FromEmail string `toml:"from_email"`
}

// Path returns the path to the config file.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "atlas", "config.toml"), nil
}

// Load reads config from ~/.config/atlas/config.toml.
// Any ATLAS_* environment variables override the file values.
func Load() (*Config, error) {
	cfg := &Config{
		IMAPPort: 993,
		SMTPPort: 587,
	}

	path, err := Path()
	if err != nil {
		return nil, fmt.Errorf("resolving config path: %w", err)
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no config file found at %s — run 'atlas setup' to create one", path)
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	// Environment variable overrides
	if v := os.Getenv("ATLAS_IMAP_HOST"); v != "" {
		cfg.IMAPHost = v
	}
	if v := os.Getenv("ATLAS_IMAP_PORT"); v != "" {
		if _, err := fmt.Sscan(v, &cfg.IMAPPort); err != nil {
			return nil, fmt.Errorf("invalid ATLAS_IMAP_PORT: %w", err)
		}
	}
	if v := os.Getenv("ATLAS_IMAP_USER"); v != "" {
		cfg.IMAPUser = v
	}
	if v := os.Getenv("ATLAS_IMAP_PASS"); v != "" {
		cfg.IMAPPass = v
	}
	if v := os.Getenv("ATLAS_SMTP_HOST"); v != "" {
		cfg.SMTPHost = v
	}
	if v := os.Getenv("ATLAS_SMTP_PORT"); v != "" {
		if _, err := fmt.Sscan(v, &cfg.SMTPPort); err != nil {
			return nil, fmt.Errorf("invalid ATLAS_SMTP_PORT: %w", err)
		}
	}
	if v := os.Getenv("ATLAS_SMTP_USER"); v != "" {
		cfg.SMTPUser = v
	}
	if v := os.Getenv("ATLAS_SMTP_PASS"); v != "" {
		cfg.SMTPPass = v
	}
	if v := os.Getenv("ATLAS_FROM_NAME"); v != "" {
		cfg.FromName = v
	}
	if v := os.Getenv("ATLAS_FROM_EMAIL"); v != "" {
		cfg.FromEmail = v
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes cfg to the config file, creating the directory if needed.
func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func validate(cfg *Config) error {
	switch {
	case cfg.IMAPHost == "":
		return errors.New("imap_host is required")
	case cfg.IMAPUser == "":
		return errors.New("imap_user is required")
	case cfg.IMAPPass == "":
		return errors.New("imap_pass is required")
	case cfg.SMTPHost == "":
		return errors.New("smtp_host is required")
	case cfg.SMTPUser == "":
		return errors.New("smtp_user is required")
	case cfg.SMTPPass == "":
		return errors.New("smtp_pass is required")
	case cfg.FromEmail == "":
		return errors.New("from_email is required")
	}
	return nil
}
