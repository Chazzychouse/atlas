package mail

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"user@example.com", true},
		{"User Name <user@example.com>", true},
		{"a@b.co", true},
		{"user+tag@example.com", true},
		{"", false},
		{"   ", false},
		{"notanemail", false},
		{"@example.com", false},
		{"user@", false},
		{"<>", false},
		{"user@@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ValidateEmail(tt.input)
			if got != tt.valid {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestFormatAddress(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"", "user@example.com", "user@example.com"},
		{"Alice", "alice@example.com", "Alice <alice@example.com>"},
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
