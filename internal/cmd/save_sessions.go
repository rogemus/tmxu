package cmd

import (
	"flag"
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
	"github.com/rogemus/tmxu/internal/utils"
)

var _ Cmd = (*SaveSessions)(nil)

type SaveSessions struct {
	ios IOs
	cfg config.Config
}

// NewSaveSessions builds a SaveSessionsCmd from the app config.
func NewSaveSessions(ios IOs, cfg config.Config) *SaveSessions {
	return &SaveSessions{ios: ios, cfg: cfg}
}

var saveSessionsInfo = Info{
	Command:   "save-sessions",
	Aliases:   []string{"save", "s"},
	Arg:       "",
	DescShort: "Save tmux sessions",
	DescLong:  "Captures all running tmux sessions including windows, panes, and layouts. Saves to ~/.config/tmxu/tmux-sessions.json.",
	Examples: []string{
		"tmxu save-sessions",
		"tmxu save",
		"tmux s",
	},
}

func (c *SaveSessions) Flags() *flag.FlagSet { return nil }
func (c *SaveSessions) Info() Info           { return saveSessionsInfo }
func (c *SaveSessions) Run(args []string) error {
	if !utils.Confirm("Save all tmux sessions?", c.ios.In, c.ios.Err) {
		fmt.Fprintln(c.ios.Err, "Aborted")
		return nil
	}

	_, err := tmux.Snapshot()
	if err != nil {
		return err
	}

	// TODO: save to file, the path is in c.cfg.Path
	// err = saveSessionsFile(tSessions)
	// if err != nil {
	// 	return fmt.Errorf("Unable to save tmux sessions to file in ~/.config/tmux \n")
	// }
	//
	// fmt.Printf("Tmux sessions saved at ~%s%s \n", configDir, sessionFile)
	return nil
}
