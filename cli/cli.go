package cli

import (
	"fmt"
	"os"
	"slices"
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
	c.newCmd(listTemplatesCmd)
	c.newCmd(saveTemplateCmd)
	c.newCmd(deleteTemplateCmd)
	c.newCmd(versionCmd)

	return &c
}

func (c *cli) help(cmdName string) {
	fmt.Println("Tmux utilities for managing sessions with save/restore capabilities")
	fmt.Println("")

	if slices.Contains(c.cmdsOrder, cmdName) {
		c.cmds[cmdName].helpLong()
		return
	}

	c.listAllCommands()

	fmt.Println("")
	fmt.Println("Use `tmxu help [command]` to get detailed information about a specific command.")
}

func (c *cli) listAllCommands() {
	var d [][]string

	for _, key := range c.cmdsOrder {
		cmd := c.cmds[key]

		d = append(d, []string{
			cmd.Command, cmd.Arg, cmd.DescShort,
		})
	}

	d = append(d, []string{"help", "[command]", "Display help information"})
	renderTable(d)
}

func (c *cli) newCmd(cmd Cmd) {
	c.cmds[cmd.Command] = cmd
	c.cmdsOrder = append(c.cmdsOrder, cmd.Command)
}

func (c *cli) Run() {
	if len(os.Args) < 2 {
		c.help("")
		os.Exit(0)
	}

	if len(os.Args) == 2 && os.Args[1] == "help" {
		c.help("")
		os.Exit(0)
	}

	if len(os.Args) == 3 && os.Args[1] == "help" {
		c.help(os.Args[2])
		os.Exit(0)
	}

	cmdName := os.Args[1]
	if cmd, ok := c.cmds[cmdName]; ok {
		if err := cmd.Run(); err != nil {
			fmt.Printf("%s", err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Println("Invalid command ")
		c.help("")
		os.Exit(1)
	}
}
