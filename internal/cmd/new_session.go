package cmd

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = (*NewSession)(nil)

var (
	ErrCannotParseNewSessionCmdFlags = fmt.Errorf("%w: cannot parse new session cmd flags", ErrUsage)
	ErrSessionNameNotProvided        = fmt.Errorf("%w: session name not provided", ErrUsage)
	ErrUnexpectedArgument            = fmt.Errorf("%w: unexpected argument, flags go before the session name", ErrUsage)
	ErrNoTemplate                    = fmt.Errorf("%w: template not found", ErrUsage)
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
	DescShort: "Create a new session, optionally from a template",
	DescLong:  "Creates a new detached tmux session. With -templ the session gets the windows and panes of a saved template. -path sets the start directory (relative paths are made absolute); it defaults to the current directory. Flags go before the session name. Run `tmxu attach <sessionName>` to use the new session.",
	Arg:       "<sessionName>",
	Examples: []string{
		"tmxu new sessionName",
		"tmxu ns sessionName",
		"tmxu new-session sessionName",
		"tmxu new-session -path /tmp/app sessionName",
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
		if strings.TrimSpace(c.templ) == "" {
			return ErrTemplateNameNotProvided
		}

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
		return fmt.Errorf("%w: %q", ErrNoTemplate, c.templ)
	}

	data, err := c.cfg.ShowTemplate(c.templ)
	if err != nil {
		return err
	}

	snapshot, err := tmux.LoadSnapshot(data)
	if err != nil {
		return err
	}

	snapshot, err = tmux.DesanitizeTemplSnapshot(snapshot, name)
	if err != nil {
		return err
	}
	err = tmux.NewSessionFromTempl(name, snapshot, c.path)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.ios.Out, "Session: created!\nRun `tmxu attach %s` in order to use newly created session\n", name)
	return err
}
