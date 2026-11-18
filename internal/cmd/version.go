package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = (*Version)(nil)

var (
	ErrCannotParseVersionCmdFlags = fmt.Errorf("%w: cannot parse version cmd flags", ErrUsage)
)

type Version struct {
	ios IOs
	cfg config.Config

	fs   *flag.FlagSet
	json bool
}

func NewVersion(ios IOs, cfg config.Config) *Version {
	cmd := &Version{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("version", flag.ContinueOnError),
	}

	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.json, "json", false, "Print JSON instead of text")

	return cmd
}

var versionInfo = Info{
	Command:   "version",
	Aliases:   []string{"v"},
	Arg:       "",
	DescShort: "Display app version information",
	DescLong:  "Shows the current tmxu version. With -json prints {\"version\":\"...\"} on one line.",
	Examples: []string{
		"tmxu version",
		"tmxu v",
		"tmxu version -json",
	},
}

func (c *Version) Flags() *flag.FlagSet { return c.fs }
func (c *Version) Info() Info           { return versionInfo }
func (c *Version) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseVersionCmdFlags, err)
	}

	if c.json {
		return printJSON(c.ios.Out, map[string]string{"version": c.cfg.Version})
	}

	_, err = fmt.Fprintf(c.ios.Out, "Current tmxu version %s\n", c.cfg.Version)
	return err
}
