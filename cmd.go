package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type cmd struct {
	command   string
	alias     string
	arg       string
	descShort string
	run       func()
}

func (c cmd) helpShort() {
	fmt.Printf(" %10s %8s    %s \n", c.command, c.arg, c.descShort)
}

var listCmd = cmd{
	command: "list",
	// TODO: suppoer alias
	alias:     "ls",
	descShort: "List all active sessions in tmux",
	run: func() {
		ls, err := ListSessions()
		if err != nil {
			LogError("Unable to list all tmux sessions")
		}

		fmt.Println("Available sessions")
		for _, s := range ls {
			parts := strings.Split(s, " ")
			fmt.Printf("%4s %15s \n", parts[0], parts[1])
		}
	},
}

var attachCmd = cmd{
	command:   "attach",
	descShort: "Attach to running tmux session",
	arg:       "[NAME]",
	run: func() {
		if len(os.Args) < 3 {
			LogError("No session name provided. Provide tmux session name you want attach to")
		}

		err := AttachToSession(os.Args[2])
		if err != nil {
			LogError("Unable to attach to tmux session: %s", os.Args[2])
		}
	},
}

var helpCmd = cmd{
	command:   "help",
	descShort: "Display help information",
	run: func() {
		attachCmd.helpShort()
		listCmd.helpShort()
		saveCmd.helpShort()
		restoreCmd.helpShort()
		fmt.Printf(" %10s %8s    %s \n", "help", "", "Display help information")
	},
}

var saveCmd = cmd{
	command:   "save",
	descShort: "Save tmux sessions",
	run: func() {
		var tSessions []tSession

		ls, err := ListSessions()
		if err != nil {
			LogError("Unable to list all tmux sessions")
		}

		for _, s := range ls {
			ts, err := newTSession(s)
			if err != nil {
				LogError("Unable to create tSession: %s", ts.Name)
			}

			lw, err := ListWindows(ts.Name)
			if err != nil {
				LogError("Unable to list windows for session: %s", ts.Name)
			}

			for _, w := range lw {
				tw, err := newTWindow(w, ts.Name)
				if err != nil {
					LogError("Unable to create tWindow: %s", tw.Name)
				}

				lp, err := ListPanes(tw.SessionWindow)
				if err != nil {
					LogError("Unable to list panes for window: %s \n", tw.SessionWindow)
				}

				for _, p := range lp {
					tp, err := newTPane(p)
					if err != nil {
						LogError("Unable to create tPane: %s", tp.Name)
					}

					tw.Panes = append(tw.Panes, tp)
				}
				ts.Windows = append(ts.Windows, tw)
			}
			tSessions = append(tSessions, ts)
		}

		err = saveFile(tSessions)
		if err != nil {
			LogError("Unable to save tmux sessions to file in ~/.config/tmux")
		}

		LogInfo("Tmux sessions saved at ~/.config/tmux")
	},
}

var restoreCmd = cmd{
	command:   "restore",
	descShort: "Restore tmux sessions",
	run: func() {
		sessions, err := loadFile()
		if err != nil {
			LogError("Unable to load sessionf from session file")
		}

		for _, session := range sessions {
			hs, err := HasSession(session.Name)
			if err != nil {
				LogError("Unable to create session: %s", session.Name)
			}

			if hs {
				continue
			}

			// for _, window := range session.Windows {
			//
			// 	// rename first window
			// 	// tmux rename-window -t worklog:1 editor
			//
			// 	// create new window
			// 	// tmux new-window -t session_name:window_index -n window_name
			// }
		}

		fmt.Printf("[%v]\n", sessions)
	},
}
