package cmd

import (
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = Version{}

type Version struct {
	version string
}

func NewVersion(cfg config.Config) Version {
	return Version{version: cfg.Version}
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
	fmt.Printf("Current tmxu version %s \n", c.version)
	return nil
}
