package composer

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Send     key.Binding
	NextField key.Binding
	PrevField key.Binding
	Cancel   key.Binding
}

// DefaultHelpBindings returns all key bindings for the help overlay.
func DefaultHelpBindings() []key.Binding {
	km := DefaultKeyMap()
	return []key.Binding{km.Send, km.NextField, km.PrevField, km.Cancel}
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Send: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "send"),
		),
		NextField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}
