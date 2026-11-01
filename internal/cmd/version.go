package cmd

import (
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = Version{}

// Version prints the tmxu version it was constructed with.
type Version struct {
	version string
}

// NewVersion builds a VersionCmd from the app config.
func NewVersion(cfg config.Config) Version {
	return Version{version: cfg.Version}
}

func (c Version) Command() string { return "version" }

func (c Version) Aliases() []string { return []string{"v"} }

func (c Version) Arg() string { return "" }

func (c Version) DescShort() string { return "Display app version information" }

func (c Version) DescLong() string { return "Shows the current tmxu version" }

func (c Version) Flags() [][]string { return nil }

func (c Version) Examples() []string {
	return []string{
		"tmxu version",
		"tmxu v",
	}
}

func (c Version) Run() error {
	fmt.Printf("Current tmxu version %s \n", c.version)
	return nil
}
