package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Cmd struct {
	Command   string
	Arg       string
	DescShort string
	Run       func() error
}

func (c Cmd) helpShort() {
	fmt.Printf(" %10s %8s    %s \n", c.Command, c.Arg, c.DescShort)
}

var listCmd = Cmd{
	Command:   "list",
	DescShort: "List all active sessions in tmux",
	Run: func() error {
		ls, err := ListSessions()
		if err != nil {
			return fmt.Errorf("Unable to list all tmux sessions")
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
	Arg:       "[NAME]",
	Run: func() error {
		if len(os.Args) < 3 {
			return fmt.Errorf("No session name provided. Provide tmux session name you want attach to")
		}

		err := AttachToSession(os.Args[2])
		if err != nil {
			return fmt.Errorf("Unable to attach to tmux session: %s \n", os.Args[2])
		}

		return nil
	},
}

var helpCmd = Cmd{
	Command:   "help",
	DescShort: "Display help information",
	Run: func() error {
		fmt.Println("Tmux utilities for managing sessions with save/restore capabilities")
		fmt.Println("")

		attachCmd.helpShort()
		listCmd.helpShort()
		saveCmd.helpShort()
		restoreCmd.helpShort()
		versionCmd.helpShort()
		fmt.Printf(" %10s %8s    %s \n", "help", "", "Display help information")
		return nil
	},
}

var versionCmd = Cmd{
	Command:   "version",
	DescShort: "Display app version information",
	Run: func() error {
		fmt.Printf("tmxu version %s", version)
		return nil
	},
}

var saveCmd = Cmd{
	Command:   "save",
	DescShort: "Save tmux sessions",
	Run: func() error {
		var tSessions []tSession

		ls, err := ListSessions()
		if err != nil {
			return fmt.Errorf("Unable to list all tmux sessions")
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

		err = saveFile(tSessions)
		if err != nil {
			return fmt.Errorf("unable to save tmux sessions to file in ~/.config/tmux \n")
		}

		fmt.Printf("Tmux sessions saved at ~%s%s \n", configDir, sessionFile)
		return nil
	},
}

var restoreCmd = Cmd{
	Command:   "restore",
	DescShort: "Restore tmux sessions",
	Run: func() error {
		sessions, err := loadFile()
		if err != nil {
			return fmt.Errorf("unable to load session from session file \n")
		}

		for _, s := range sessions {
			err := NewSession(s)
			if errors.Is(err, errorSessionExists) {
				fmt.Printf("Session already exist: %s \n", s.Name)
				continue
			} else if err != nil {
				return fmt.Errorf("unable to create session: %s \n", s.Name)
			}

			for _, window := range s.Windows {
				if err := NewWindow(window); err != nil {
					return fmt.Errorf("unable to create window: %s \n", window.SessionWindow)
				}

				for _, pane := range window.Panes {
					if err := NewPane(pane); err != nil {
						return fmt.Errorf("unable to create pane: %s \n", pane.Name)
					}
				}
			}
		}

		return nil
	},
}
