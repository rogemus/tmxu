package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type app struct {
	cmds map[string]cmd
}

func newApp() *app {
	cmds := make(map[string]cmd)

	app := &app{
		cmds: cmds,
	}

	return app
}

func (a *app) newCmd(c cmd) {
	a.cmds[c.command] = c
}

func (a *app) run() {
	if len(os.Args) < 2 {
		for _, c := range a.cmds {
			c.helpShort()
		}

		os.Exit(1)
	}

	c, ok := a.cmds[os.Args[1]]
	if ok {
		c.run()
	} else {
		fmt.Println("invalid command ")

		for _, c := range a.cmds {
			c.helpShort()
		}
	}
}

func main() {
	// restoreCmd := flag.NewFlagSet("restore", flag.ExitOnError)
	// saveCmd := flag.NewFlagSet("save", flag.ExitOnError)
	// newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	// attachCmd := flag.NewFlagSet("attach", flag.ExitOnError)
	// helpCmd := flag.NewFlagSet("help", flag.ExitOnError)

	app := newApp()

	listCmd := cmd{
		command: "list",
		// TODO: suppoer alias
		alias:     "ls",
		descShort: "List all active sessions in tmux",
		run: func() {
			cmd, err := exec.Command("tmux", "ls").Output()
			if err != nil {
				// TODO print nice error
				print(err)
				os.Exit(1)
			}

			sessions := strings.Split(strings.TrimSpace(string(cmd)), "\n")

			fmt.Printf("Available sessions: \n")
			for _, session := range sessions {
				parts := strings.Split(session, " ")
				fmt.Printf(" %15s %s windows \n", parts[0], parts[1])
			}

			os.Exit(0)
		},
	}

	attachCmd := cmd{
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

			os.Exit(0)
		},
	}

	helpCmd := cmd{
		command:   "help",
		descShort: "Display help information",
		run: func() {
			for _, c := range app.cmds {
				c.helpShort()
			}
		},
	}

	app.newCmd(attachCmd)
	app.newCmd(listCmd)
	app.newCmd(helpCmd)

	app.run()
}
