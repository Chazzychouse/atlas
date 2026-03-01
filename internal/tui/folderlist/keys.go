package folderlist

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Select key.Binding
	New    key.Binding
	Rename key.Binding
	Delete key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new folder"),
		),
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
	}
}

func DefaultHelpBindings() []key.Binding {
	km := DefaultKeyMap()
	return []key.Binding{km.Select, km.New, km.Rename, km.Delete}
}
