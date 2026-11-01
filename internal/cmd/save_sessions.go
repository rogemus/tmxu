package cmd

import (
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
	"github.com/rogemus/tmxu/internal/utils"
)

var _ Cmd = SaveSessions{}

// SaveSessions captures the running tmux sessions and saves them to the
// path from the app config.
type SaveSessions struct {
	cfg config.Config
}

// NewSaveSessions builds a SaveSessionsCmd from the app config.
func NewSaveSessions(cfg config.Config) SaveSessions {
	return SaveSessions{cfg: cfg}
}

func (c SaveSessions) Command() string { return "save-sessions" }

func (c SaveSessions) Aliases() []string { return []string{"save", "s"} }

func (c SaveSessions) Arg() string { return "" }

func (c SaveSessions) DescShort() string { return "Save tmux sessions" }

func (c SaveSessions) DescLong() string {
	return "Captures all running tmux sessions including windows, panes, and layouts. Saves to ~/.config/tmxu/tmux-sessions.json."
}

func (c SaveSessions) Flags() [][]string { return nil }

func (c SaveSessions) Examples() []string {
	return []string{
		"tmxu save-sessions",
		"tmxu save",
		"tmux s",
	}
}

func (c SaveSessions) Run() error {
	if !utils.Confirm("Save all tmux sessions?") {
		fmt.Print("Aborted")
		return nil
	}

	_, err := tmux.GetTmuxData()
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
