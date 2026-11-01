package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = AttachSession{}

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
	var menuMode bool
	fs := flag.NewFlagSet("attach-session", flag.ContinueOnError)
	fs.BoolVar(&menuMode, "menu", false, "Use interactive menu to select session")

	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("unable to parse flags \n")
	}

	// err := tmux.AttachToSession(sessionName)
	// if err != nil {
	// 	return err
	// }

	return nil

}
