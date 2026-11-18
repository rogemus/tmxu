// Package cli holds the command registry. It looks up a command by name or
// alias, dispatches it with the remaining arguments, and prints help.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/rogemus/tmxu/internal/cmd"
	"github.com/rogemus/tmxu/internal/config"
)

var (
	ErrUnknownCommand = fmt.Errorf("%w: unknown command", cmd.ErrUsage)
)

type Cli struct {
	cmds      map[string]cmd.Cmd
	cmdsOrder []string
	ios       cmd.IOs
}

func (c *Cli) Run(args []string) error {
	if len(args) == 0 {
		return c.helpAll(c.ios.Out)
	}

	name, rest := args[0], args[1:]
	if name == "help" || name == "-h" || name == "--help" {
		return c.help(c.ios.Out, rest)
	}

	command, ok := c.cmds[name]
	if !ok {
		return c.unknownCmdError(name)
	}

	if wantsHelp(rest) {
		return c.helpCmd(c.ios.Out, command)
	}

	err := command.Run(rest)
	if errors.Is(err, cmd.ErrUsage) {
		return fmt.Errorf("%w\nUsage: %s", err, command.Info().Usage())
	}

	return err
}

var cmdHelpTmpl = template.Must(
	template.New("cmdHelp").
		Funcs(template.FuncMap{"join": strings.Join}).
		Parse(`{{.Info.Command}} - {{.Info.DescShort}}
{{- if .Info.Aliases}}

ALIASES:
  {{join .Info.Aliases ", "}}
{{- end}}
{{- if .Info.DescLong}}

DESCRIPTION:
  {{.Info.DescLong}}
{{- end}}

USAGE:
  {{.Info.Usage}}{{if .Flags}} [flags]{{end}}
{{- if .Flags}}

FLAGS:
{{- range .Flags}}
  -{{.Name}}      {{.Desc}}
{{- end}}
{{- end}}
{{- if .Info.Examples}}

EXAMPLES:
{{- range .Info.Examples}}
  {{.}}
{{- end}}
{{- end}}
`))

func (c *Cli) help(w io.Writer, args []string) error {
	if len(args) == 0 {
		return c.helpAll(w)
	}

	command, ok := c.cmds[args[0]]
	if !ok {
		return c.unknownCmdError(args[0])
	}

	return c.helpCmd(w, command)
}

func (c *Cli) helpCmd(w io.Writer, command cmd.Cmd) error {
	var data = struct {
		Info  cmd.Info
		Flags []cmd.Flag
	}{
		Info:  command.Info(),
		Flags: make([]cmd.Flag, 0),
	}

	command.Flags().VisitAll(func(f *flag.Flag) {
		data.Flags = append(data.Flags, cmd.Flag{Name: f.Name, Desc: f.Usage})
	})

	return cmdHelpTmpl.Execute(w, data)
}

func (c *Cli) helpAll(w io.Writer) error {
	var bw strings.Builder
	fmt.Fprintln(&bw, "Tmux utilities for managing sessions with save/restore capabilities")
	fmt.Fprintln(&bw)

	c.listAllCommands(&bw)

	fmt.Fprintln(&bw)
	fmt.Fprintln(&bw, "Use `tmxu help [command]` to get detailed information about a specific command")

	_, err := io.WriteString(w, bw.String())
	return err
}

func (c *Cli) listAllCommands(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, name := range c.cmdsOrder {
		info := c.cmds[name].Info()
		label := info.Command

		if aliases := info.Aliases; len(aliases) > 0 {
			label += " (" + strings.Join(aliases, ", ") + ")"
		}

		fmt.Fprintf(tw, "  %s\t%s\t%s\n", label, info.Arg, info.DescShort)
	}

	tw.Flush()
}

func (c *Cli) unknownCmdError(name string) error {
	_ = c.helpAll(c.ios.Err) // best effort; the unknown-command error matters more
	return fmt.Errorf("%w: %q", ErrUnknownCommand, name)
}

func wantsHelp(args []string) bool {
	for _, a := range args {
		if a == "--" {
			return false
		}
		if a == "-h" || a == "-help" || a == "--help" {
			return true
		}
	}

	return false
}

func (c *Cli) newCmd(command cmd.Cmd) {
	info := command.Info()
	c.register(info.Command, command)
	c.cmdsOrder = append(c.cmdsOrder, info.Command)

	for _, alias := range info.Aliases {
		c.register(alias, command)
	}
}

func (c *Cli) register(name string, command cmd.Cmd) {
	if name == "help" || name == "-h" || name == "--help" {
		panic(fmt.Sprintf("command name %q is reserved for help", name))
	}

	if other, ok := c.cmds[name]; ok {
		panic(fmt.Sprintf("command name %q already used by: %s", name, other.Info().Command))
	}

	c.cmds[name] = command
}

func New(cfg config.Config, ios cmd.IOs) *Cli {
	c := Cli{
		cmds:      make(map[string]cmd.Cmd),
		cmdsOrder: make([]string, 0),
		ios:       ios,
	}

	aCmd := cmd.NewAttachSession(ios)
	c.newCmd(aCmd)
	lsCmd := cmd.NewListSessions(ios)
	c.newCmd(lsCmd)
	ssCmd := cmd.NewSaveSessions(ios, cfg)
	c.newCmd(ssCmd)
	rsCmd := cmd.NewRestoreSessions(ios, cfg)
	c.newCmd(rsCmd)
	nsCmd := cmd.NewNewSession(ios, cfg)
	c.newCmd(nsCmd)
	dCmd := cmd.NewDump(ios)
	c.newCmd(dCmd)
	vCmd := cmd.NewVersion(ios, cfg)
	c.newCmd(vCmd)
	pCmd := cmd.NewPaths(ios, cfg)
	c.newCmd(pCmd)
	ltCmd := cmd.NewListTemplates(ios, cfg)
	c.newCmd(ltCmd)
	stCmd := cmd.NewSaveTemplate(ios, cfg)
	c.newCmd(stCmd)
	dtCmd := cmd.NewDeleteTemplate(ios, cfg)
	c.newCmd(dtCmd)
	shtCmd := cmd.NewShowTemplate(ios, cfg)
	c.newCmd(shtCmd)

	return &c
}
