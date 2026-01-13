package main

import (
	"fmt"
	"os"
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
		helpCmd := a.cmds["help"]
		helpCmd.run()

		os.Exit(1)
	}

	c, ok := a.cmds[os.Args[1]]
	if ok {
		c.run()
	} else {
		fmt.Println("invalid command ")
		helpCmd := a.cmds["help"]
		helpCmd.run()

		os.Exit(1)
	}
}

func main() {
	app := newApp()

	app.newCmd(attachCmd)
	app.newCmd(listCmd)
	app.newCmd(saveCmd)
	app.newCmd(helpCmd)

	app.run()
}
