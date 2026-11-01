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
		info := c.Info()
		name := info.Command

		if seen[name] {
			continue
		} else {
			seen[name] = true
		}

		if aliases := info.Aliases; len(aliases) > 0 {
			name += " (" + strings.Join(aliases, ", ") + ")"
		}

		fmt.Fprintf(tw, "  %s\t%s\t%s\n", name, info.Arg, info.DescShort)
	}
	tw.Flush()
}

func (c *Cli) newCmd(cmd cmd.Cmd) {
	info := cmd.Info()
	c.cmds[info.Command] = cmd
	c.cmdsOrder = append(c.cmdsOrder, info.Command)

	for _, alias := range info.Aliases {
		c.cmds[alias] = cmd
	}
}

func New(cfg config.Config) *Cli {
	c := Cli{
		cmds:      make(map[string]cmd.Cmd),
		cmdsOrder: make([]string, 0),
	}

	lsCmd := cmd.NewListSessions()
	c.newCmd(lsCmd)
	ssCmd := cmd.NewSaveSessions(cfg)
	c.newCmd(ssCmd)
	vCmd := cmd.NewVersion(cfg)
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
