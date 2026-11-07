package cmd

import (
	"flag"
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = (*Version)(nil)

type Version struct {
	ios IOs
	cfg config.Config
}

func NewVersion(ios IOs, cfg config.Config) *Version {
	return &Version{ios: ios, cfg: cfg}
}

var versionInfo = Info{
	Command:   "version",
	Aliases:   []string{"v"},
	Arg:       "",
	DescShort: "Display app version information",
	DescLong:  "Shows the current tmxu version",
	Examples: []string{
		"tmxu version",
		"tmxu v",
	},
}

func (c *Version) Flags() *flag.FlagSet { return nil }
func (c *Version) Info() Info           { return versionInfo }
func (c *Version) Run(args []string) error {
	_, err := fmt.Fprintf(c.ios.Out, "Current tmxu version %s\n", c.cfg.Version)
	return err
}
