package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/rogemus/tmxu/internal/cmd"
	"github.com/rogemus/tmxu/internal/config"
)

var (
	ErrUnknownCommand = errors.New("unknown command")
)

type Cli struct {
	cmds      map[string]cmd.Cmd
	cmdsOrder []string
	ios       cmd.IOs
}

func (c Cli) Run(args []string) error {
	if len(args) == 0 {
		return c.helpAll(c.ios.Out)
	}

	name, rest := args[0], args[1:]
	if name == "help" {
		return c.help(c.ios.Out, rest)
	}

	command, ok := c.cmds[name]
	if !ok {
		return c.unknownCmdError(name)
	}

	return command.Run(rest)
}

var cmdHelpTmpl = template.Must(
	template.New("cmdHelp").
		Funcs(template.FuncMap{"join": strings.Join}).
		Parse(`{{.Command}} - {{.DescShort}}
{{- if .Aliases}}

ALIASES:
  {{join .Aliases ", "}}
{{- end}}
{{- if .DescLong}}

DESCRIPTION:
  {{.DescLong}}
{{- end}}

USAGE:
  tmxu {{.Command}}{{with .Arg}} {{.}}{{end}}{{if .Flags}} [flags]{{end}}
{{- if .Flags}}

FLAGS:
{{- range .Flags}}
  -{{index . 0}}      {{index . 1}}
{{- end}}
{{- end}}
{{- if .Examples}}

EXAMPLES:
{{- range .Examples}}
  {{.}}
{{- end}}
{{- end}}
`))

func (c Cli) help(w io.Writer, args []string) error {
	if len(args) == 0 {
		return c.helpAll(c.ios.Err)
	}

	command, ok := c.cmds[args[0]]
	if !ok {
		return c.unknownCmdError(args[0])
	}

	return c.helpCmd(w, command)
}

func (c Cli) helpCmd(w io.Writer, command cmd.Cmd) error {
	return cmdHelpTmpl.Execute(w, command.Info())
}

func (c Cli) helpAll(w io.Writer) error {
	var bw strings.Builder
	fmt.Fprintln(&bw, "Tmux utilities for managing sessions with save/restore capabilities")
	fmt.Fprintln(&bw)

	c.listAllCommands(&bw)

	fmt.Fprintln(&bw)
	fmt.Fprintf(&bw, "Use `tmxu help [command]` to get detailed information about a specific command\n")

	_, err := io.WriteString(w, bw.String())
	return err
}

func (c Cli) listAllCommands(w io.Writer) {
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

func (c Cli) unknownCmdError(name string) error {
	_ = c.helpAll(c.ios.Err) // best effort; the unknown-command error matters more
	return fmt.Errorf("%w: %q", ErrUnknownCommand, name)
}

func (c *Cli) newCmd(command cmd.Cmd) {
	info := command.Info()
	c.cmds[info.Command] = command
	c.cmdsOrder = append(c.cmdsOrder, info.Command)

	for _, alias := range info.Aliases {
		c.cmds[alias] = command
	}
}

func New(cfg config.Config, ios cmd.IOs) *Cli {
	c := Cli{
		cmds:      make(map[string]cmd.Cmd),
		cmdsOrder: make([]string, 0),
		ios:       ios,
	}

	lsCmd := cmd.NewListSessions(cfg, ios)
	c.newCmd(lsCmd)
	ssCmd := cmd.NewSaveSessions(cfg, ios)
	c.newCmd(ssCmd)
	vCmd := cmd.NewVersion(cfg, ios)
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
