package cli

import (
	"fmt"
	"os"
)

var version string

type cli struct {
	cmds      map[string]Cmd
	cmdsOrder []string
}

func NewCli(v string) *cli {
	version = v

	c := cli{
		cmds:      make(map[string]Cmd),
		cmdsOrder: make([]string, 0),
	}

	c.newCmd(attachCmd)
	c.newCmd(listCmd)
	c.newCmd(saveCmd)
	c.newCmd(restoreCmd)
	c.newCmd(versionCmd)

	return &c
}

func (c *cli) help() {
	fmt.Println("Tmux utilities for managing sessions with save/restore capabilities")
	fmt.Println("")

	for _, key := range c.cmdsOrder {
		c.cmds[key].helpShort()
	}

	fmt.Printf(" %10s %8s    %s \n", "help", "", "Display help information")
}

func (c *cli) newCmd(cmd Cmd) {
	c.cmds[cmd.Command] = cmd
	c.cmdsOrder = append(c.cmdsOrder, cmd.Command)
}

func (c *cli) Run() {
	if len(os.Args) < 2 {
		c.help()
		os.Exit(0)
	}

	if cmd, ok := c.cmds[os.Args[1]]; ok {
		if err := cmd.Run(); err != nil {
			fmt.Printf("%s", err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Println("Invalid command ")
		c.help()
		os.Exit(1)
	}
}
