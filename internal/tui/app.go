package tui

import (
	"strings"
	"time"

	"github.com/chazzychouse/atlas/internal/config"
	"github.com/chazzychouse/atlas/internal/contacts"
	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui/statusbar"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewFactory creates views by ID. Set by the app to avoid circular imports.
type ViewFactory func(id ViewID, data PushViewMsg) View

// App is the root Bubble Tea model.
type App struct {
	cfg           *config.Config
	imap          *mail.IMAPClient
	smtp          *mail.SMTPClient
	contacts      *contacts.Manager
	nav           *NavStack
	statusbar     statusbar.Model
	keys          GlobalKeyMap
	factory       ViewFactory
	width         int
	height        int
	showFolder    bool
	folderFocused bool
	folderView    View
	helpView      View
	showHelp      bool
	ready         bool
	statusGen     uint64 // incremented each time a status is set; used to expire auto-clears
}

// NewApp creates the root application model.
func NewApp(cfg *config.Config, imap *mail.IMAPClient, smtp *mail.SMTPClient, ctcts *contacts.Manager) *App {
	return &App{
		cfg:       cfg,
		imap:      imap,
		smtp:      smtp,
		contacts:  ctcts,
		nav:       NewNavStack(),
		statusbar: statusbar.New(),
		keys:      DefaultGlobalKeyMap(),
	}
}

// SetFactory sets the view factory. Called from cmd/root.go after all packages are imported.
func (a *App) SetFactory(f ViewFactory) {
	a.factory = f
}

// Config returns the app config.
func (a *App) Config() *config.Config { return a.cfg }

// IMAP returns the IMAP client.
func (a *App) IMAP() *mail.IMAPClient { return a.imap }

// SMTP returns the SMTP client.
func (a *App) SMTP() *mail.SMTPClient { return a.smtp }

// Nav returns the navigation stack.
func (a *App) Nav() *NavStack { return a.nav }

// Width returns the terminal width.
func (a *App) Width() int { return a.width }

// Height returns the available height for the main view.
func (a *App) Height() int { return a.height - 2 } // subtract help line + status bar

// FolderWidth returns the width of the folder sidebar.
func (a *App) FolderWidth() int {
	if a.showFolder {
		return 25
	}
	return 0
}

// MainWidth returns the width of the main content area.
func (a *App) MainWidth() int {
	return a.width - a.FolderWidth()
}

// Init initializes the app by connecting to IMAP and loading the envelope list.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.connectAndLoad(),
	)
}

func (a *App) connectAndLoad() tea.Cmd {
	return func() tea.Msg {
		if err := a.imap.Connect(); err != nil {
			return ErrMsg{Err: err}
		}
		return StatusMsg{Text: "Connected"}
	}
}

// scanSentContacts fetches recent envelopes from the sent folder for contact extraction.
func (a *App) scanSentContacts() tea.Cmd {
	return func() tea.Msg {
		// Try common sent-folder names in priority order.
		folders, err := a.imap.ListFolders()
		if err != nil {
			return nil
		}

		sentFolder := findSentFolder(folders)
		if sentFolder == "" {
			return nil
		}

		envs, err := a.imap.FetchRecentFromFolder(sentFolder, 200)
		if err != nil {
			return nil
		}
		return ContactsSyncedMsg{Envelopes: envs}
	}
}

// findSentFolder picks the best-matching sent folder from a folder list.
func findSentFolder(folders []mail.Folder) string {
	// Prefer exact Gmail name, then fall back to name-contains match.
	priority := []string{
		"[Gmail]/Sent Mail",
		"Sent Mail",
		"Sent Messages",
		"Sent Items",
		"Sent",
	}
	byName := make(map[string]bool, len(folders))
	for _, f := range folders {
		byName[f.Name] = true
	}
	for _, name := range priority {
		if byName[name] {
			return name
		}
	}
	// Fallback: case-insensitive substring search.
	for _, f := range folders {
		lower := strings.ToLower(f.Name)
		if strings.Contains(lower, "sent") {
			return f.Name
		}
	}
	return ""
}

// Update handles messages for the root app.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.statusbar.SetWidth(msg.Width)
		if !a.ready {
			a.ready = true
			// Create initial envelope list view
			if a.factory != nil {
				view := a.factory(ViewEnvelopeList, PushViewMsg{})
				a.nav.Push(view)
				cmds = append(cmds, view.Init())
			}
		}
		// Propagate resize to current view
		if v := a.nav.Current(); v != nil {
			updated, cmd := v.Update(msg)
			a.nav.Replace(updated)
			cmds = append(cmds, cmd)
		}
		if a.folderView != nil {
			updated, cmd := a.folderView.Update(msg)
			a.folderView = updated
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		// Global keys first
		switch {
		case key.Matches(msg, a.keys.Quit):
			if a.showHelp {
				a.showHelp = false
				return a, nil
			}
			if a.nav.Len() <= 1 {
				a.imap.Close()
				return a, tea.Quit
			}
		case key.Matches(msg, a.keys.Help):
			a.showHelp = !a.showHelp
			if a.showHelp && a.helpView == nil && a.factory != nil {
				a.helpView = a.factory(ViewHelp, PushViewMsg{})
			}
			return a, nil
		case key.Matches(msg, a.keys.FolderList):
			a.showFolder = !a.showFolder
			if a.showFolder {
				a.folderFocused = true
				if a.folderView == nil && a.factory != nil {
					a.folderView = a.factory(ViewFolderList, PushViewMsg{})
					cmds = append(cmds, a.folderView.Init())
				}
			} else {
				a.folderFocused = false
			}
			return a, tea.Batch(cmds...)
		case key.Matches(msg, a.keys.Back):
			if a.showHelp {
				a.showHelp = false
				return a, nil
			}
			if a.folderFocused {
				a.folderFocused = false
				return a, nil
			}
			if a.nav.Len() > 1 {
				a.nav.Pop()
				return a, nil
			}
		case msg.String() == "tab" && a.showFolder:
			a.folderFocused = !a.folderFocused
			return a, nil
		}

		// Route keys to focused panel only
		if a.folderFocused && a.showFolder && a.folderView != nil {
			updated, cmd := a.folderView.Update(msg)
			a.folderView = updated
			cmds = append(cmds, cmd)
		} else if v := a.nav.Current(); v != nil {
			updated, cmd := v.Update(msg)
			a.nav.Replace(updated)
			cmds = append(cmds, cmd)
		}

	case PushViewMsg:
		if a.factory != nil {
			view := a.factory(msg.ViewID, msg)
			a.nav.Push(view)
			cmds = append(cmds, view.Init())
		}

	case PopViewMsg:
		if a.nav.Len() > 1 {
			a.nav.Pop()
		}

	case FolderSelectedMsg:
		a.showFolder = false
		a.folderFocused = false
		// Push new envelope list for selected folder
		if a.factory != nil {
			view := a.factory(ViewEnvelopeList, PushViewMsg{Folder: msg.Folder})
			a.nav.Replace(view)
			cmds = append(cmds, view.Init())
		}

	case StatusMsg:
		a.statusbar.SetStatus(msg.Text, msg.IsError)
		// After the initial IMAP connection, kick off a background sent-folder scan.
		if msg.Text == "Connected" && a.contacts != nil {
			cmds = append(cmds, a.scanSentContacts())
		}
		a.statusGen++
		if msg.Text != "" && !msg.IsError {
			gen := a.statusGen
			cmds = append(cmds, func() tea.Msg {
				time.Sleep(4 * time.Second)
				return StatusClearMsg{Gen: gen}
			})
		}

	case StatusClearMsg:
		if msg.Gen == a.statusGen {
			a.statusbar.SetStatus("", false)
		}

	case SpinnerStartMsg:
		cmds = append(cmds, a.statusbar.StartSpinner())

	case SpinnerStopMsg:
		a.statusbar.StopSpinner()

	case ErrMsg:
		a.statusbar.SetStatus("Error: "+msg.Err.Error(), true)
		a.statusbar.StopSpinner()

	case ContactsSyncedMsg:
		if a.contacts != nil {
			for _, env := range msg.Envelopes {
				a.contacts.Update(env.From, env.Date)
				for _, addr := range env.To {
					a.contacts.Update(addr, env.Date)
				}
				for _, addr := range env.Cc {
					a.contacts.Update(addr, env.Date)
				}
			}
		}

	case EnvelopesLoadedMsg:
		if msg.Err == nil && a.contacts != nil {
			for _, env := range msg.Envelopes {
				a.contacts.Update(env.From, env.Date)
				for _, addr := range env.To {
					a.contacts.Update(addr, env.Date)
				}
				for _, addr := range env.Cc {
					a.contacts.Update(addr, env.Date)
				}
			}
		}
		if v := a.nav.Current(); v != nil {
			updated, cmd := v.Update(msg)
			a.nav.Replace(updated)
			cmds = append(cmds, cmd)
		}

	default:
		// Pass to current view
		if v := a.nav.Current(); v != nil {
			updated, cmd := v.Update(msg)
			a.nav.Replace(updated)
			cmds = append(cmds, cmd)
		}

		// Pass to folder view (for async results like FolderCreatedMsg, etc.)
		if a.folderView != nil {
			updated, cmd := a.folderView.Update(msg)
			a.folderView = updated
			cmds = append(cmds, cmd)
		}

		// Update spinner
		sb, cmd := a.statusbar.Update(msg)
		a.statusbar = sb
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

// View renders the full application.
func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	var mainContent string
	if a.showHelp && a.helpView != nil {
		mainContent = a.helpView.View()
	} else if v := a.nav.Current(); v != nil {
		mainContent = v.View()
	}

	var body string
	if a.showFolder && a.folderView != nil {
		sidebarStyle := lipgloss.NewStyle().
			Width(a.FolderWidth()).
			Height(a.Height())
		if a.folderFocused {
			sidebarStyle = sidebarStyle.
				BorderLeft(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Width(a.FolderWidth() - 1)
		}
		sidebar := sidebarStyle.Render(a.folderView.View())

		main := lipgloss.NewStyle().
			Width(a.MainWidth()).
			Height(a.Height()).
			Render(mainContent)

		body = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
	} else {
		body = lipgloss.NewStyle().
			Width(a.width).
			Height(a.Height()).
			Render(mainContent)
	}

	return lipgloss.JoinVertical(lipgloss.Left, body, a.renderHelpLine(), a.statusbar.View())
}

// renderHelpLine builds a single-line help bar from global + current view keybindings.
func (a *App) renderHelpLine() string {
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorBright)
	descStyle := lipgloss.NewStyle().Foreground(ColorDim)
	sep := descStyle.Render(" │ ")

	// Collect bindings: global keys first, then current view's keys.
	var bindings []key.Binding
	bindings = append(bindings, a.keys.Help, a.keys.Quit, a.keys.FolderList, a.keys.Back)

	if a.showHelp && a.helpView != nil {
		bindings = append(bindings, a.helpView.ShortHelp()...)
	} else if a.folderFocused && a.showFolder && a.folderView != nil {
		bindings = append(bindings, a.folderView.ShortHelp()...)
	} else if v := a.nav.Current(); v != nil {
		bindings = append(bindings, v.ShortHelp()...)
	}

	var parts []string
	for _, b := range bindings {
		if !b.Enabled() {
			continue
		}
		h := b.Help()
		if h.Key == "" {
			continue
		}
		parts = append(parts, keyStyle.Render(h.Key)+" "+descStyle.Render(h.Desc))
	}

	line := " " + strings.Join(parts, sep)

	return lipgloss.NewStyle().
		Width(a.width).
		Background(lipgloss.Color("#222222")).
		Render(line)
}
