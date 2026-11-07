package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = (*AttachSession)(nil)

var (
	ErrCannotParseAttachCmdFlags = fmt.Errorf("%w: cannot parse attach cmd flags", ErrUsage)
)

type AttachSession struct {
	ios IOs

	fs   *flag.FlagSet
	menu bool
}

func NewAttachSession(ios IOs) *AttachSession {
	cmd := &AttachSession{
		ios: ios,
		fs:  flag.NewFlagSet("attach-session", flag.ContinueOnError),
	}

	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.menu, "menu", false, "Use interactive menu to select session")
	return cmd
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

func (c *AttachSession) Flags() *flag.FlagSet { return nil }
func (c *AttachSession) Info() Info           { return attachSessionInfo }
func (c *AttachSession) Run(args []string) error {

	// TODO: implement fzf style menu
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseAttachCmdFlags, err)
	}

	if len(c.fs.Args()) == 0 {
		return ErrSessionNameNotProvided
	}

	return tmux.AttachToSession(c.fs.Arg(0))
}
