package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = (*Paths)(nil)

var (
	ErrCannotParsePathsCmdFlags = fmt.Errorf("%w: cannot parse paths cmd flags", ErrUsage)
)

type Paths struct {
	ios IOs
	cfg config.Config

	fs   *flag.FlagSet
	json bool
}

func NewPaths(ios IOs, cfg config.Config) *Paths {
	cmd := &Paths{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("paths", flag.ContinueOnError),
	}

	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.json, "json", false, "Print JSON instead of text")

	return cmd
}

var pathsInfo = Info{
	Command:   "paths",
	Arg:       "",
	DescShort: "Show where tmxu keeps its files",
	DescLong:  "Prints the config directory, the templates directory and the saved sessions file. Nothing is created. With -json prints {\"config\":\"...\",\"templates\":\"...\",\"sessions\":\"...\"} on one line.",
	Examples: []string{
		"tmxu paths",
		"tmxu paths -json",
	},
}

func (c *Paths) Flags() *flag.FlagSet { return c.fs }
func (c *Paths) Info() Info           { return pathsInfo }
func (c *Paths) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParsePathsCmdFlags, err)
	}

	if c.json {
		return printJSON(c.ios.Out, map[string]string{
			"config":    c.cfg.AppDirPath,
			"templates": c.cfg.TemplatesDir(),
			"sessions":  c.cfg.SessionsFile(),
		})
	}

	_, err = fmt.Fprintf(c.ios.Out, "config:    %s\ntemplates: %s\nsessions:  %s\n",
		c.cfg.AppDirPath, c.cfg.TemplatesDir(), c.cfg.SessionsFile())
	return err
}
