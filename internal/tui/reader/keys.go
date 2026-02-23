package reader

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Reply    key.Binding
	ReplyAll key.Binding
	Forward  key.Binding
	Delete   key.Binding
}

// DefaultHelpBindings returns all key bindings for the help overlay.
func DefaultHelpBindings() []key.Binding {
	km := DefaultKeyMap()
	return []key.Binding{km.Reply, km.ReplyAll, km.Forward, km.Delete}
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Reply: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reply"),
		),
		ReplyAll: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "reply all"),
		),
		Forward: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "forward"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
	}
}
