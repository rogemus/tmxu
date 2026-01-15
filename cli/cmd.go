package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type cmd struct {
	command   string
	arg       string
	descShort string
	run       func() error
}

func (c cmd) helpShort() {
	fmt.Printf(" %10s %8s    %s \n", c.command, c.arg, c.descShort)
}

var listCmd = cmd{
	command:   "list",
	descShort: "List all active sessions in tmux",
	run: func() error {
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

var attachCmd = cmd{
	command:   "attach",
	descShort: "Attach to running tmux session",
	arg:       "[NAME]",
	run: func() error {
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

var helpCmd = cmd{
	command:   "help",
	descShort: "Display help information",
	run: func() error {
		attachCmd.helpShort()
		listCmd.helpShort()
		saveCmd.helpShort()
		restoreCmd.helpShort()
		fmt.Printf(" %10s %8s    %s \n", "help", "", "Display help information")
		return nil
	},
}

var saveCmd = cmd{
	command:   "save",
	descShort: "Save tmux sessions",
	run: func() error {
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

		fmt.Println("Tmux sessions saved at ~/.config/tmux")
		return nil
	},
}

var restoreCmd = cmd{
	command:   "restore",
	descShort: "Restore tmux sessions",
	run: func() error {
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
