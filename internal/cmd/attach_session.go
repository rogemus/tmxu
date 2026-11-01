package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = AttachSession{}

var (
	ErrCannotParseAttachCmdFlags = fmt.Errorf("%w: cannot parse attach cmd flags", ErrUsage)
	ErrSessionNameNotProvided    = fmt.Errorf("%w: session name not provided", ErrUsage)
)

type AttachSession struct {
	cfg config.Config
	ios IOs
}

func NewAttachSession(cfg config.Config, ios IOs) AttachSession {
	return AttachSession{cfg, ios}
}

var attachSessionInfo = Info{
	Command:   "attach-session",
	Aliases:   []string{"attach", "a"},
	DescShort: "Attach to running tmux session",
	DescLong:  "Connects to an existing tmux session by name. The session must already be running.",
	Arg:       "[sessionName]",
	Examples: []string{
		"tmxu attach-session mysession",
		"tmux attach-session -menu",
		"tmxu attach mysession",
		"tmxu a mysession",
	},
}

func (c AttachSession) Info() Info { return attachSessionInfo }
func (c AttachSession) Run(args []string) error {

	// TODO: implement fzf style menu
	var menuMode bool
	fs := flag.NewFlagSet("attach-session", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.BoolVar(&menuMode, "menu", false, "Use interactive menu to select session")

	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseAttachCmdFlags, err)
	}

	if len(fs.Args()) == 0 {
		return ErrSessionNameNotProvided
	}

	return tmux.AttachToSession(fs.Arg(0))
}
