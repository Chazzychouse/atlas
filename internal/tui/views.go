package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewID identifies different views in the application.
type ViewID int

const (
	ViewEnvelopeList ViewID = iota
	ViewReader
	ViewComposer
	ViewFolderList
	ViewHelp
)

// View is the interface that all sub-views must implement.
type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
	ShortHelp() []key.Binding
}

// HelpSection is a named group of key bindings for the help overlay.
type HelpSection struct {
	Title    string
	Bindings []key.Binding
}
