package folderlist

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Select key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select folder"),
		),
	}
}
