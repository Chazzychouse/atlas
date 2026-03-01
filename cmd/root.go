package cmd

import (
	"fmt"
	"os"

	"github.com/chazzychouse/atlas/internal/config"
	"github.com/chazzychouse/atlas/internal/mail"
	"github.com/chazzychouse/atlas/internal/tui"
	"github.com/chazzychouse/atlas/internal/tui/composer"
	"github.com/chazzychouse/atlas/internal/tui/envelopelist"
	"github.com/chazzychouse/atlas/internal/tui/folderlist"
	"github.com/chazzychouse/atlas/internal/tui/help"
	"github.com/chazzychouse/atlas/internal/tui/reader"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "atlas",
	Short: "A TUI email client",
	Long:  "Atlas is a terminal-based email client with native IMAP and SMTP support.",
	RunE:  runTUI,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	imapClient := mail.NewIMAPClient(cfg)
	smtpClient := mail.NewSMTPClient(cfg)

	app := tui.NewApp(cfg, imapClient, smtpClient)

	// Wire up the ViewFactory — this is the only place that imports all sub-packages.
	app.SetFactory(func(id tui.ViewID, data tui.PushViewMsg) tui.View {
		switch id {
		case tui.ViewEnvelopeList:
			folder := data.Folder
			if folder == "" {
				folder = "INBOX"
			}
			return envelopelist.New(imapClient, folder, app.MainWidth(), app.Height())

		case tui.ViewReader:
			return reader.New(imapClient, data.EnvelopeUID, app.MainWidth(), app.Height())

		case tui.ViewComposer:
			c := composer.New(cfg, smtpClient, app.MainWidth(), app.Height())
			if data.ReplyTo != nil {
				c.Prefill(data.ReplyTo, data.ReplyAll, data.Forward)
			}
			return c

		case tui.ViewFolderList:
			return folderlist.New(imapClient, app.FolderWidth(), app.Height())

		case tui.ViewHelp:
			globalKeys := tui.DefaultGlobalKeyMap()
			sections := []tui.HelpSection{
				{Title: "Global", Bindings: []key.Binding{
					globalKeys.Quit, globalKeys.Help, globalKeys.FolderList, globalKeys.Back,
				}},
				{Title: "Envelope List", Bindings: envelopelist.DefaultHelpBindings()},
				{Title: "Message Reader", Bindings: reader.DefaultHelpBindings()},
				{Title: "Composer", Bindings: composer.DefaultHelpBindings()},
			}
			return help.New(sections, app.MainWidth(), app.Height())

		default:
			return envelopelist.New(imapClient, "INBOX", app.MainWidth(), app.Height())
		}
	})

	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return err
	}

	return nil
}
