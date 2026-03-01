package contacts

import (
	"encoding/json"
	"net/mail"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// DefaultConfigDir returns the atlas config directory path.
func DefaultConfigDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.Getenv("HOME") + "/.config"
	}
	return filepath.Join(base, "atlas")
}

// Contact represents a known email contact.
type Contact struct {
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	LastSeen time.Time `json:"last_seen"`
	Count    int       `json:"count"`
}

// Formatted returns "Name <email>" or bare email if no name.
func (c Contact) Formatted() string {
	if c.Name != "" {
		return c.Name + " <" + c.Email + ">"
	}
	return c.Email
}

// Manager stores and queries email contacts.
type Manager struct {
	path     string
	mu       sync.RWMutex
	contacts map[string]*Contact // keyed by lowercase email
}

// New creates a Manager that persists to configDir/contacts.json.
func New(configDir string) *Manager {
	return &Manager{
		path:     filepath.Join(configDir, "contacts.json"),
		contacts: make(map[string]*Contact),
	}
}

// Load reads contacts from disk. Silently succeeds if file doesn't exist.
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var list []Contact
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range list {
		key := strings.ToLower(list[i].Email)
		if key != "" {
			m.contacts[key] = &list[i]
		}
	}
	return nil
}

func (m *Manager) save() {
	m.mu.RLock()
	list := make([]Contact, 0, len(m.contacts))
	for _, c := range m.contacts {
		list = append(list, *c)
	}
	m.mu.RUnlock()

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(m.path), 0o750)
	_ = os.WriteFile(m.path, data, 0o600)
}

// Update adds or refreshes a contact from an address string.
// addr may be "Name <email>" or bare "email". when is the email date.
func (m *Manager) Update(addr string, when time.Time) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return
	}

	parsed, err := mail.ParseAddress(addr)
	if err != nil {
		return
	}

	key := strings.ToLower(parsed.Address)
	if key == "" {
		return
	}

	m.mu.Lock()
	c, ok := m.contacts[key]
	if !ok {
		c = &Contact{Email: parsed.Address}
		m.contacts[key] = c
	}
	if parsed.Name != "" {
		c.Name = parsed.Name
	}
	if when.After(c.LastSeen) {
		c.LastSeen = when
	}
	c.Count++
	m.mu.Unlock()

	go m.save()
}

// Search returns contacts matching query (case-insensitive substring of name or email),
// sorted by LastSeen descending. Returns nil if query is empty.
func (m *Manager) Search(query string) []Contact {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	lower := strings.ToLower(query)

	m.mu.RLock()
	var results []Contact
	for _, c := range m.contacts {
		if strings.Contains(strings.ToLower(c.Email), lower) ||
			strings.Contains(strings.ToLower(c.Name), lower) {
			results = append(results, *c)
		}
	}
	m.mu.RUnlock()

	sort.Slice(results, func(i, j int) bool {
		return results[i].LastSeen.After(results[j].LastSeen)
	})
	return results
}
