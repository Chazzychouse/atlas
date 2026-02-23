package envelopelist

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Open       key.Binding
	Compose    key.Binding
	Delete     key.Binding
	Move       key.Binding
	ToggleRead key.Binding
	Reply      key.Binding
	ReplyAll   key.Binding
	Forward    key.Binding
	NextPage   key.Binding
	PrevPage   key.Binding
	Refresh    key.Binding
	Select     key.Binding
	SelectAll  key.Binding
}

// DefaultHelpBindings returns all key bindings for the help overlay.
func DefaultHelpBindings() []key.Binding {
	km := DefaultKeyMap()
	return []key.Binding{
		km.Open, km.Compose, km.Delete, km.ToggleRead,
		km.Reply, km.ReplyAll, km.Forward,
		km.NextPage, km.PrevPage, km.Refresh,
		km.Select, km.SelectAll,
	}
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Open: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open message"),
		),
		Compose: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "compose"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Move: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "move"),
		),
		ToggleRead: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "toggle read/unread"),
		),
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
		NextPage: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next page"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "prev page"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("V"),
			key.WithHelp("V", "select all"),
		),
	}
}
