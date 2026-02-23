package tui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeyMap defines key bindings available in all views.
type GlobalKeyMap struct {
	Quit       key.Binding
	Help       key.Binding
	FolderList key.Binding
	Back       key.Binding
}

// DefaultGlobalKeyMap returns the default global key bindings.
func DefaultGlobalKeyMap() GlobalKeyMap {
	return GlobalKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		FolderList: key.NewBinding(
			key.WithKeys("F"),
			key.WithHelp("F", "toggle folders"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}
