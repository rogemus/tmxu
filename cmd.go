package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type cmd struct {
	command   string
	alias     string
	arg       string
	descShort string
	run       func()
}

type tSession struct {
	Order   int16     `json:"order"`
	Name    string    `json:"name"`
	Windows []tWindow `json:"windows"`
}

type tWindow struct {
	Order  int16   `json:"order"`
	Name   string  `json:"name"`
	Layout string  `json:"layout"`
	Panes  []tPane `json:"panes"`
}

type tPane struct {
	Order int16  `json:"order"`
	Name  string `json:"name"`
	Path  string `json:"path"`
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
		cmd, err := exec.Command("tmux", "ls").Output()

		if err != nil {
			fmt.Println("unable to list tmux session")
			os.Exit(1)
		}

		sessions := strings.Split(strings.TrimSpace(string(cmd)), "\n")

		fmt.Printf("Available sessions: \n")
		for _, session := range sessions {
			parts := strings.Split(session, " ")
			fmt.Printf(" %15s %s windows \n", parts[0], parts[1])
		}
	},
}

var attachCmd = cmd{
	command:   "attach",
	descShort: "Attach to running tmux session",
	arg:       "[NAME]",
	run: func() {
		if len(os.Args) < 3 {
			fmt.Println("provide tmux session name you want attach to")
			os.Exit(1)
		}

		name := os.Args[2]
		err := exec.Command("tmux", "attach", "-t", name).Run()
		if err != nil {
			fmt.Printf("unable to attach to tmux session: %s \n", name)
			os.Exit(1)
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
		fmt.Printf(" %10s %8s    %s \n", "help", "", "Display help information")
	},
}

var saveCmd = cmd{
	command:   "save",
	descShort: "Save tmux sessions",
	run: func() {
		sessionsCmd, err := exec.Command("tmux", "list-sessions", "-F", "#{session_id} #{session_name}").Output()

		if err != nil {
			fmt.Println("unable to list active tmux sessions")
			os.Exit(1)
		}

		var tSessions []tSession

		// list all sessions
		sessions := strings.SplitSeq(strings.TrimSpace(string(sessionsCmd)), "\n")
		for session := range sessions {
			parts := strings.Split(session, " ")
			order, err := strconv.Atoi(strings.TrimPrefix(parts[0], "$"))
			if err != nil {
				fmt.Printf("unable to parse order for session: %s", parts[1])
				os.Exit(1)
			}

			sessionName := parts[1]
			s := tSession{
				Order: int16(order),
				Name:  sessionName,
			}

			windowsCmd, err := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_index} #{window_name} #{window_layout}").Output()
			if err != nil {
				fmt.Printf("unable to list session windows: %s \n", sessionName)
				os.Exit(1)
			}

			// list all windows for session
			windows := strings.SplitSeq(strings.TrimSpace(string(windowsCmd)), "\n")
			for window := range windows {
				parts := strings.Split(window, " ")
				order, err := strconv.Atoi(parts[0])
				if err != nil {
					fmt.Printf("unable to parse order for window: %s", parts[1])
					os.Exit(1)
				}

				w := tWindow{
					Order:  int16(order),
					Name:   parts[1],
					Layout: parts[2],
				}

				sessionWindow := fmt.Sprintf("%s:%s", sessionName, parts[0])
				panesCmd, err := exec.Command("tmux", "list-panes", "-t", sessionWindow, "-F", "#{pane_index} #{pane_title} #{pane_current_path}").Output()
				if err != nil {
					fmt.Printf("unable to list panes for window: %s \n", sessionWindow)
					os.Exit(1)
				}

				// list all panes for window in session
				panes := strings.SplitSeq(strings.TrimSpace(string(panesCmd)), "\n")
				for pane := range panes {
					parts := strings.Split(pane, " ")
					order, err := strconv.Atoi(parts[0])
					if err != nil {
						fmt.Printf("unable to parse order for pane: %s", parts[1])
						os.Exit(1)
					}

					p := tPane{
						Order: int16(order),
						Name:  parts[1],
						Path:  parts[2],
					}

					w.Panes = append(w.Panes, p)
				}
				s.Windows = append(s.Windows, w)
			}
			tSessions = append(tSessions, s)
		}

		saveFile(tSessions)
	},
}
