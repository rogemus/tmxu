package cmd

import (
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = Version{}

type Version struct {
	cfg config.Config
	ios IOs
}

func NewVersion(cfg config.Config, io IOs) Version {
	return Version{cfg, io}
}

var versionInfo = Info{
	Command:   "version",
	Aliases:   []string{"v"},
	Arg:       "",
	DescShort: "Display app version information",
	DescLong:  "Shows the current tmxu version",
	Flags:     nil,
	Examples: []string{
		"tmxu version",
		"tmxu v",
	},
}

func (c Version) Info() Info { return versionInfo }
func (c Version) Run(args []string) error {
	_, err := fmt.Fprintf(c.ios.Out, "Current tmxu version %s\n", c.cfg.Version)
	return err
}
