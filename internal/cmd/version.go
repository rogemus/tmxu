package cmd

import (
	"fmt"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = VersionCmd{}

// VersionCmd prints the tmxu version it was constructed with.
type VersionCmd struct {
	Version string
}

// NewVersionCmd builds a VersionCmd from the app config.
func NewVersionCmd(cfg config.Cfg) VersionCmd {
	return VersionCmd{Version: cfg.Version}
}

func (c VersionCmd) Command() string { return "version" }

func (c VersionCmd) Aliases() []string { return []string{"v"} }

func (c VersionCmd) Arg() string { return "" }

func (c VersionCmd) DescShort() string { return "Display app version information" }

func (c VersionCmd) DescLong() string { return "Shows the current tmxu version" }

func (c VersionCmd) Flags() [][]string { return nil }

func (c VersionCmd) Examples() []string {
	return []string{
		"tmxu version",
		"tmxu v",
	}
}

func (c VersionCmd) Run() error {
	fmt.Printf("Current tmxu version %s \n", c.Version)
	return nil
}
