package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/rogemus/tmxu/internal/cmd"
	"github.com/rogemus/tmxu/internal/config"
)

type Cli struct {
	cmds      map[string]cmd.Cmd
	cmdsOrder []string
}

func (c *Cli) Run() {
	c.help("")
}

func (c Cli) help(cmdName string) {
	w := os.Stdout
	fmt.Fprintln(w, "Tmux utilities for managing sessions with save/restore capabilities")
	fmt.Fprintln(w)

	c.listAllCommands(w)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "Use `tmxu help [command]` to get detailed information about a specific command.")
}

func (c Cli) listAllCommands(w io.Writer) {
	seen := make(map[string]bool)

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, c := range c.cmds {
		name := c.Command()

		if seen[name] {
			continue
		} else {
			seen[name] = true
		}

		if aliases := c.Aliases(); len(aliases) > 0 {
			name += " (" + strings.Join(aliases, ", ") + ")"
		}

		fmt.Fprintf(tw, "  %s\t%s\t%s\n", name, c.Arg(), c.DescShort())
	}
	tw.Flush()
}

func (c *Cli) newCmd(cmd cmd.Cmd) {
	c.cmds[cmd.Command()] = cmd
	c.cmdsOrder = append(c.cmdsOrder, cmd.Command())

	for _, alias := range cmd.Aliases() {
		c.cmds[alias] = cmd
	}
}

func New(cfg config.Cfg) *Cli {
	c := Cli{
		cmds:      make(map[string]cmd.Cmd),
		cmdsOrder: make([]string, 0),
	}

	lsCmd := cmd.NewListSessionsCmd()
	c.newCmd(lsCmd)
	ssCmd := cmd.NewSaveSessionsCmd(cfg)
	c.newCmd(ssCmd)
	vCmd := cmd.NewVersionCmd(cfg)
	c.newCmd(vCmd)

	// c.newCmd(newSessionCmd)
	// c.newCmd(attachSessionCmd)
	// c.newCmd(listSessionsCmd)
	// c.newCmd(saveSessionsCmd)
	// c.newCmd(restoreSessionsCmd)
	// c.newCmd(listTemplatesCmd)
	// c.newCmd(saveTemplateCmd)
	// c.newCmd(deleteTemplateCmd)
	// c.newCmd(versionCmd)

	return &c
}
