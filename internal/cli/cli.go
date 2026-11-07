package cli

import (
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

	if c.wantsHelp(rest) {
		return c.helpCmd(c.ios.Out, command)
	}

	return command.Run(rest)
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
  tmxu {{.Info.Command}}{{with .Info.Arg}} {{.}}{{end}}{{if .Flags}} [flags]{{end}}
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
		return c.helpAll(c.ios.Out)
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

	fs := command.Flags()
	if fs != nil {
		fs.VisitAll(func(f *flag.Flag) {
			data.Flags = append(data.Flags, cmd.Flag{Name: f.Name, Desc: f.Usage})
		})
	}

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

func (c *Cli) wantsHelp(args []string) bool {
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
	c.cmds[info.Command] = command
	c.cmdsOrder = append(c.cmdsOrder, info.Command)

	for _, alias := range info.Aliases {
		com, ok := c.cmds[alias]
		if ok {
			// NOTE: internal error, update cmds config
			msg := fmt.Sprintf("alias already used by: %s\n", com.Info().Command)
			panic(msg)
		}
		c.cmds[alias] = command
	}
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
	dCmd := cmd.NewDump(ios)
	c.newCmd(dCmd)
	vCmd := cmd.NewVersion(ios, cfg)
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
