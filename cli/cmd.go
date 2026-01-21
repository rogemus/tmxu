package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Cmd struct {
	Command   string
	Arg       string
	DescShort string
	DescLong  string
	Flags     [][]string
	Examples  []string
	Run       func() error
}

func (c Cmd) helpLong() {
	fmt.Printf("%s - %s\n\n", c.Command, c.DescShort)

	if c.DescLong != "" {
		fmt.Printf("DESCRIPTION:\n  %s\n\n", c.DescLong)
	}

	usage := fmt.Sprintf("  tmxu %s", c.Command)
	if c.Arg != "" {
		usage += " " + c.Arg
	}
	if len(c.Flags) > 0 {
		usage += " [flags]"
	}
	fmt.Printf("USAGE:\n%s\n\n", usage)

	if len(c.Flags) > 0 {
		fmt.Println("FLAGS:")
		for _, f := range c.Flags {
			fmt.Printf("  -%s      %s\n", f[0], f[1])
		}
	}
}

var listCmd = Cmd{
	Command:   "list",
	DescShort: "List all active sessions in tmux",
	DescLong:  "Displays all currently running tmux sessions with their IDs and names.",
	Examples: []string{
		"tmxu list",
	},
	Run: func() error {
		ls, err := ListSessions()
		if err != nil {
			return fmt.Errorf("Unable to list all tmux sessions \n")
		}

		fmt.Println("Available sessions")
		for _, s := range ls {
			parts := strings.Split(s, " ")
			fmt.Printf("%4s %15s \n", parts[0], parts[1])
		}

		return nil
	},
}

var attachCmd = Cmd{
	Command:   "attach",
	DescShort: "Attach to running tmux session",
	DescLong:  "Connects to an existing tmux session by name. The session must already be running.",
	Arg:       "[sessionName]",
	Examples: []string{
		"tmxu attach mysession",
	},
	Run: func() error {
		if len(os.Args) < 3 {
			return fmt.Errorf("No session name provided. Provide tmux session name you want attach to \n")
		}

		err := AttachToSession(os.Args[2])
		if err != nil {
			return fmt.Errorf("Unable to attach to tmux session: %s \n", os.Args[2])
		}

		return nil
	},
}

var versionCmd = Cmd{
	Command:   "version",
	DescShort: "Display app version information",
	DescLong:  "Shows the current tmxu version and checks GitHub for newer releases.",
	Examples: []string{
		"tmxu version",
	},
	Run: func() error {
		ghVersion, err := getNewestVersion()
		if err != nil {
			fmt.Printf("tmxu version %s \n", version)
			return fmt.Errorf("Unable to check for the newest version on Github \n")
		}

		sv := newSemVer(version)
		if sv.original == ghVersion.original {
			fmt.Printf("tmxu version %s \n", version)
			return nil
		}

		fmt.Println("A new version of the tmxu is available!")
		fmt.Println("Please run the following command to update:")
		fmt.Printf("  go install github.com/rogemus/tmxu@%s\n\n", ghVersion.original)
		fmt.Printf("Current tmxu version %s \n", version)
		return nil
	},
}

var saveCmd = Cmd{
	Command:   "save",
	DescShort: "Save tmux sessions",
	DescLong:  "Captures all running tmux sessions including windows, panes, and layouts. Saves to ~/.config/tmxu/tmux-sessions.json.",
	Examples: []string{
		"tmxu save",
	},
	Run: func() error {
		if !confirm("Save all tmux sessions?") {
			fmt.Println("Aborted.")
			return nil
		}

		var tSessions []tSession

		ls, err := ListSessions()
		if err != nil {
			return fmt.Errorf("Unable to list all tmux sessions \n")
		}

		for _, s := range ls {
			ts, err := newTSession(s)
			if err != nil {
				return fmt.Errorf("Unable to create tSession: %s \n", ts.Name)
			}

			lw, err := ListWindows(ts.Name)
			if err != nil {
				return fmt.Errorf("Unable to list windows for session: %s \n", ts.Name)
			}

			for _, w := range lw {
				tw, err := newTWindow(w, ts.Name)
				if err != nil {
					return fmt.Errorf("Unable to create tWindow: %s \n", tw.Name)
				}

				lp, err := ListPanes(tw.SessionWindow)
				if err != nil {
					return fmt.Errorf("Unable to list panes for window: %s \n", tw.SessionWindow)
				}

				for _, p := range lp {
					tp, err := newTPane(p, tw.SessionName, tw.SessionWindow)
					if err != nil {
						return fmt.Errorf("Unable to create tPane: %s \n", tp.Name)
					}

					tw.Panes = append(tw.Panes, tp)
				}
				ts.Windows = append(ts.Windows, tw)
			}
			tSessions = append(tSessions, ts)
		}

		err = saveSessionsFile(tSessions)
		if err != nil {
			return fmt.Errorf("Unable to save tmux sessions to file in ~/.config/tmux \n")
		}

		fmt.Printf("Tmux sessions saved at ~%s%s \n", configDir, sessionFile)
		return nil
	},
}

var restoreCmd = Cmd{
	Command:   "restore",
	DescShort: "Restore tmux sessions",
	DescLong:  "Recreates tmux sessions from ~/.config/tmxu/tmux-sessions.json. Skips sessions that already exist.",
	Flags: [][]string{
		{"force", "override existing sessions while restoring"},
	},
	Examples: []string{
		"tmxu restore",
		"tmux restore -force",
	},
	Run: func() error {
		var force bool
		fs := flag.NewFlagSet("restore", flag.ContinueOnError)
		fs.BoolVar(&force, "force", false, "override existing sessions while restoring")

		if !confirm("Restore tmux sessions from saved file?") {
			fmt.Println("Aborted.")
			return nil
		}

		sessions, err := loadSessionsFile()
		if err != nil {
			return fmt.Errorf("Unable to load session from session file \n")
		}

		for _, s := range sessions {
			err := NewSession(s, force)
			if errors.Is(err, errorSessionExists) {
				fmt.Printf("Session already exist: %s \n", s.Name)
				continue
			} else if err != nil {
				return fmt.Errorf("Unable to create session: %s \n", s.Name)
			}

			for _, window := range s.Windows {
				if err := NewWindow(window); err != nil {
					return fmt.Errorf("Unable to create window: %s \n", window.SessionWindow)
				}

				for _, pane := range window.Panes {
					if err := NewPane(pane); err != nil {
						return fmt.Errorf("Unable to create pane: %s \n", pane.Name)
					}
				}
			}
		}

		return nil
	},
}

var listTemplatesCmd = Cmd{
	Command:   "list-templates",
	DescShort: "List all saved templates",
	DescLong:  "Displays all saved templates with their windows and panes. Templates are stored in ~/.config/tmxu/templates/.",
	Examples: []string{
		"tmxu list-templates",
	},
	Run: func() error {
		ts, err := listTemplates()
		if err != nil {
			return fmt.Errorf("Unable to list availabe templates in `~/.config/tmxu/templates` \n")
		}

		if len(ts) == 0 {
			fmt.Printf("No saved templates \n")
			return nil
		}

		for _, t := range ts {
			fmt.Printf("%s: %d windows \n", t.Name, len(t.Windows))
			for _, w := range t.Windows {
				fmt.Printf("  window %s: %d panes \n", w.Name, len(w.Panes))
				for _, p := range w.Panes {
					fmt.Printf("    pane: %s \n", p.Name)
				}
			}
		}

		return nil
	},
}

var saveTemplateCmd = Cmd{
	Command:   "save-template",
	DescShort: "Save session as template",
	DescLong:  "Saves a running tmux session as a reusable template. Templates are stored in ~/.config/tmxu/templates/.",
	Arg:       "[sessionName]",
	Flags: [][]string{
		{"name", "Name of the template. Defaults to session name"},
	},
	Examples: []string{
		"tmxu save-template sessionName",
		"tmxu save-template -name templateName sessionName",
	},
	Run: func() error {
		var templateName string

		fs := flag.NewFlagSet("new-session", flag.ContinueOnError)
		fs.StringVar(&templateName, "name", "", "Name of the template. Default to session name")

		if len(os.Args) < 3 {
			return fmt.Errorf("No session name provided. Provide tmux session name you want save as template \n")
		}

		if err := fs.Parse(os.Args[2:]); err != nil {
			return fmt.Errorf("Unable to read flags for cmd \n")
		}

		sessionName := fs.Arg(0)
		hs, err := HasSession(sessionName)
		if err != nil || !hs {
			return fmt.Errorf("Unable to check session: %s \n", sessionName)
		}

		if templateName == "" {
			templateName = sessionName
		}

		const TEMP_VALUE = "TEMP_VALUE"
		ts := tSession{
			Name: templateName,
		}

		lw, err := ListWindows(TEMP_VALUE)
		if err != nil {
			return fmt.Errorf("Unable to list windows for session: %s \n", ts.Name)
		}

		for _, w := range lw {
			tw, err := newTWindow(w, TEMP_VALUE)
			if err != nil {
				return fmt.Errorf("Unable to create tWindow: %s \n", tw.Name)
			}

			lp, err := ListPanes(TEMP_VALUE)
			if err != nil {
				return fmt.Errorf("Unable to list panes for window: %s \n", tw.SessionWindow)
			}

			for _, p := range lp {
				tp, err := newTPane(p, TEMP_VALUE, TEMP_VALUE)
				if err != nil {
					return fmt.Errorf("Unable to create tPane: %s \n", tp.Name)
				}

				tp.Path = TEMP_VALUE
				tw.Panes = append(tw.Panes, tp)
			}

			ts.Windows = append(ts.Windows, tw)
		}

		err = saveTemplateFile(tTemplate(ts))
		if err != nil {
			return fmt.Errorf("Unable to save session: %s as template \n", sessionName)
		}

		fmt.Printf("Templates saved at: ~/.config/tmxu/templates/%s.json \n", ts.Name)
		return nil
	},
}

var deleteTemplateCmd = Cmd{
	Command:   "delete-template",
	DescShort: "Delete saved template",
	DescLong:  "Removes a template file from ~/.config/tmxu/templates/.",
	Arg:       "[templateName]",
	Examples: []string{
		"tmxu delete-template templateName",
	},
	Run: func() error {
		if len(os.Args) < 3 {
			return fmt.Errorf("No template name provided. Provide template name you want to delete \n")
		}

		templateName := os.Args[2]
		if err := deleteTemplateFile(templateName); err != nil {
			return fmt.Errorf("Unable to delete template: `~/.config/tmxu/templates/%s.json` \n", templateName)
		}

		fmt.Printf("Template deleted: `~/.config/tmxu/templates/%s.json` \n", templateName)
		return nil
	},
}

var newSessionCmd = Cmd{
	Command:   "new-session",
	DescShort: "Create new session base on the template",
	DescLong:  "Creates a new tmux session, optionally based on a saved template.",
	Arg:       "[sessionName]",
	Flags: [][]string{
		{"path", "Initial path for all panes. Defaults to current directory"},
		{"templ", "Template to create new session based on"},
	},
	Examples: []string{
		"tmxu new-session sessionName",
		"tmxu new-session -templ templateName sessionName",
		"tmxu new-session -path /tmp/app -templ templateName sessionName",
	},
	Run: func() error {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("Cannot get pwd for current dir \n")
		}

		var (
			path  string
			templ string
		)

		fs := flag.NewFlagSet("new-session", flag.ContinueOnError)
		fs.StringVar(&path, "path", pwd, "Initial path for all panes. Default to pwd")
		fs.StringVar(&templ, "templ", "", "Template to create new session base on")

		if err := fs.Parse(os.Args[2:]); err != nil {
			return fmt.Errorf("Unable to read cmd options \n")
		}

		sessionName := fs.Arg(0)
		if templ == "" {
			t := tSession{
				Name: sessionName,
			}

			err := NewSession(t, false)
			if errors.Is(err, errorSessionExists) {
				return fmt.Errorf("Session already exist: %s \n", t.Name)
			} else if err != nil {
				return fmt.Errorf("Unable to create session: %s \n", t.Name)
			}

			return nil
		}

		t, err := loadTemplateFile(templ)
		t.Name = sessionName
		if err != nil {
			return fmt.Errorf("Unable to read template file: %s \n", sessionName)
		}

		err = NewSession(t, false)
		if errors.Is(err, errorSessionExists) {
			return fmt.Errorf("Session already exist: %s \n", t.Name)
		} else if err != nil {
			return fmt.Errorf("Unable to create session: %s \n", t.Name)
		}

		for i, window := range t.Windows {
			window.SessionName = sessionName
			window.SessionWindow = fmt.Sprintf("%s:%d", sessionName, i+1)

			if err := NewWindow(window); err != nil {
				return fmt.Errorf("Unable to create window: %s \n", window.SessionWindow)
			}

			for _, pane := range window.Panes {
				pane.Path = path
				pane.SessionName = sessionName
				pane.SessionWindow = window.SessionWindow

				if err := NewPane(pane); err != nil {
					return fmt.Errorf("Unable to create pane: %s \n", pane.Name)
				}
			}
		}

		fmt.Printf("Session: %s created! \nRun `tmxu attach %s` in order to use newly created session \n", sessionName, sessionName)
		return nil
	},
}
