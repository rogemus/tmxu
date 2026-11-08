package cmd

import (
	"flag"
	"io"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = (*RestoreSessions)(nil)

type RestoreSessions struct {
	ios IOs
	cfg config.Config

	fs    *flag.FlagSet
	force bool
}

func NewRestoreSessions(ios IOs, cfg config.Config) *RestoreSessions {
	cmd := &RestoreSessions{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("restore-sessions", flag.ContinueOnError),
	}
	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.force, "force", false, "override existing sessions while restoring")
	return cmd
}

var restoreSessionsInfo = Info{
	Command:   "restore-sessions",
	Aliases:   []string{"restore", "r"},
	DescShort: "Restore tmux sessions",
	DescLong:  "Recreates tmux sessions from ~/.config/tmxu/tmux-sessions.json. Skips sessions that already exist.",
	Examples: []string{
		"tmxu restore-sessions",
		"tmxu restore",
		"tmxu r",
		"tmxu restore -force",
	},
}

func (c *RestoreSessions) Flags() *flag.FlagSet { return c.fs }
func (c *RestoreSessions) Info() Info           { return restoreSessionsInfo }
func (c *RestoreSessions) Run(args []string) error {
	return nil
}
