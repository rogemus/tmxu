package cli

import (
	"fmt"
	"os"
)

var version string

type cli struct {
	cmds map[string]Cmd
}

func NewCli(v string) *cli {
	version = v

	c := cli{
		cmds: make(map[string]Cmd),
	}

	c.newCmd(attachCmd)
	c.newCmd(listCmd)
	c.newCmd(saveCmd)
	c.newCmd(restoreCmd)
	c.newCmd(versionCmd)
	c.newCmd(helpCmd)

	return &c
}

func (c *cli) newCmd(cmd Cmd) {
	c.cmds[cmd.Command] = cmd
}

func (c *cli) Run() {
	helpCmd := c.cmds["help"]

	if len(os.Args) < 2 {
		helpCmd.Run()
		return
		os.Exit(0)
	}

	if c, ok := c.cmds[os.Args[1]]; ok {
		if err := c.Run(); err != nil {
			fmt.Printf("%s", err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Println("Invalid command ")
		helpCmd.Run()
		os.Exit(1)
	}
}
