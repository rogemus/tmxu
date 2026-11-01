package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = AttachSession{}

var (
	ErrCannotParseAttachCmdFlags = errors.New("cannot parse attach cmd flags")
	ErrSessionNameNotProvided    = errors.New("session name not provided")
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

	if len(args) == 0 {
		return ErrSessionNameNotProvided
	}

	var menuMode bool
	fs := flag.NewFlagSet("attach-session", flag.ContinueOnError)
	fs.BoolVar(&menuMode, "menu", false, "Use interactive menu to select session")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseAttachCmdFlags, err)
	}

	err := tmux.AttachToSession(args[0])
	if err != nil {
		return err
	}

	return nil
}
