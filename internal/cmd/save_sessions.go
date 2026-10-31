package cmd

import (
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
	"github.com/rogemus/tmxu/internal/utils"
)

var _ Cmd = SaveSessionsCmd{}

// SaveSessionsCmd captures the running tmux sessions and saves them to the
// path from the app config.
type SaveSessionsCmd struct {
	cfg config.Cfg
}

// NewSaveSessionsCmd builds a SaveSessionsCmd from the app config.
func NewSaveSessionsCmd(cfg config.Cfg) SaveSessionsCmd {
	return SaveSessionsCmd{cfg: cfg}
}

func (c SaveSessionsCmd) Command() string { return "save-sessions" }

func (c SaveSessionsCmd) Aliases() []string { return []string{"save", "s"} }

func (c SaveSessionsCmd) Arg() string { return "" }

func (c SaveSessionsCmd) DescShort() string { return "Save tmux sessions" }

func (c SaveSessionsCmd) DescLong() string {
	return "Captures all running tmux sessions including windows, panes, and layouts. Saves to ~/.config/tmxu/tmux-sessions.json."
}

func (c SaveSessionsCmd) Flags() [][]string { return nil }

func (c SaveSessionsCmd) Examples() []string {
	return []string{
		"tmxu save-sessions",
		"tmxu save",
		"tmux s",
	}
}

func (c SaveSessionsCmd) Run() error {
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
