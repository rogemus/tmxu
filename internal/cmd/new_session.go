package cmd

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = NewSession{}

var (
	ErrCannotParseNewSessionCmdFlags = fmt.Errorf("%w: cannot parse new session cmd flags", ErrUsage)
	ErrSessionNameNotProvided        = fmt.Errorf("%w: session name not provided", ErrUsage)
	ErrUnexpectedArgument            = fmt.Errorf("%w: unexpected argument, flags go before the session name", ErrUsage)
)

type NewSession struct {
	ios IOs
}

func NewNewSession(io IOs) NewSession {
	return NewSession{io}
}

var newSessionInfo = Info{
	Command:   "new-session",
	Aliases:   []string{"new", "ns"},
	DescShort: "Create new session base on the template",
	DescLong:  "Creates a new tmux session, optionally based on a saved template.",
	Arg:       "[sessionName]",
	Flags: []Flag{
		{"path", "Initial path for all panes. Defaults to the current directory"},
		{"templ", "Template to create new session based on"},
	},
	Examples: []string{
		"tmxu new sessionName",
		"tmxu ns sessionName",
		"tmxu new-session sessionName",
		"tmxu new-session -templ templateName sessionName",
		"tmxu new-session -path /tmp/app -templ templateName sessionName",
	},
}

func (c NewSession) Info() Info { return newSessionInfo }
func (c NewSession) Run(args []string) error {
	var (
		path  string
		templ string
	)

	fs := flag.NewFlagSet("new-session", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&path, "path", "", "Initial path for all panes. Defaults to the current directory")
	// TODO: template support
	fs.StringVar(&templ, "templ", "", "Template to create new session based on")

	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseNewSessionCmdFlags, err)
	}

	if path != "" {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}

	name := fs.Arg(0)
	if name == "" {
		return ErrSessionNameNotProvided
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("%w: %q", ErrUnexpectedArgument, fs.Args()[1:])
	}

	err = tmux.NewSession(name, path)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.ios.Out, "Session: created!\nRun `tmxu attach %s` in order to use newly created session\n", name)
	return err
}
