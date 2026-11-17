package cmd

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = (*NewSession)(nil)

var (
	ErrCannotParseNewSessionCmdFlags = fmt.Errorf("%w: cannot parse new session cmd flags", ErrUsage)
	ErrSessionNameNotProvided        = fmt.Errorf("%w: session name not provided", ErrUsage)
	ErrUnexpectedArgument            = fmt.Errorf("%w: unexpected argument, flags go before the session name", ErrUsage)
	ErrInvalidTemplName              = fmt.Errorf("%w: invalid template name", ErrUsage)
)

type NewSession struct {
	ios IOs
	cfg config.Config

	fs    *flag.FlagSet
	path  string
	templ string
}

func NewNewSession(ios IOs, cfg config.Config) *NewSession {
	cmd := &NewSession{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("new-session", flag.ContinueOnError),
	}
	cmd.fs.SetOutput(io.Discard)
	cmd.fs.StringVar(&cmd.path, "path", "", "Initial path for all panes. Defaults to the current directory")

	// TODO: template support
	cmd.fs.StringVar(&cmd.templ, "templ", "", "Template name to create new session based on")
	return cmd
}

var newSessionInfo = Info{
	Command:   "new-session",
	Aliases:   []string{"new", "ns"},
	DescShort: "Create new session base on the template",
	DescLong:  "Creates a new tmux session, optionally based on a saved template.",
	Arg:       "[sessionName]",
	Examples: []string{
		"tmxu new sessionName",
		"tmxu ns sessionName",
		"tmxu new-session sessionName",
		"tmxu new-session -templ templateName sessionName",
		"tmxu new-session -path /tmp/app -templ templateName sessionName",
	},
}

func (c *NewSession) Flags() *flag.FlagSet { return c.fs }
func (c *NewSession) Info() Info           { return newSessionInfo }
func (c *NewSession) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseNewSessionCmdFlags, err)
	}

	if c.path != "" {
		c.path, err = filepath.Abs(c.path)
		if err != nil {
			return err
		}
	}

	name := c.fs.Arg(0)
	if name == "" {
		return ErrSessionNameNotProvided
	}
	if c.fs.NArg() > 1 {
		return fmt.Errorf("%w: %q", ErrUnexpectedArgument, c.fs.Args()[1:])
	}

	if c.templ != "" {
		return c.newSessionFromTempl(name)
	}

	return c.newSession(name)
}

func (c *NewSession) newSession(name string) error {
	err := tmux.NewSession(name, c.path)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.ios.Out, "Session: created!\nRun `tmxu attach %s` in order to use newly created session\n", name)
	return err
}

func (c *NewSession) newSessionFromTempl(name string) error {
	hasTemplate, err := c.cfg.HasTemplate(c.templ)
	if err != nil {
		return err
	}
	if !hasTemplate {
		return fmt.Errorf("%w: %w", ErrInvalidTemplName, err)
	}

	data, err := c.cfg.ShowTemplate(c.templ)
	if err != nil {
		return err
	}

	snapshot, err := tmux.LoadSnapshot(data)
	if err != nil {
		return err
	}

	snapshot = tmux.DesanitizeTemplSnapshot(snapshot, name)
	err = tmux.NewSessionFromTempl(name, snapshot, c.path)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.ios.Out, "Session: created!\nRun `tmxu attach %s` in order to use newly created session\n", name)
	return err
}
