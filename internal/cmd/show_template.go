package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = (*ShowTemplate)(nil)

var (
	ErrCannotParseShowTemplateCmdFlags = fmt.Errorf("%w: cannot parse show template cmd flags", ErrUsage)
	ErrTemplateNameNotProvided         = errors.New("template name not provided")
)

type ShowTemplate struct {
	ios IOs
	cfg config.Config

	fs   *flag.FlagSet
	json bool
}

func NewShowTemplate(ios IOs, cfg config.Config) *ShowTemplate {
	cmd := &ShowTemplate{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("show-template", flag.ContinueOnError),
	}
	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.json, "json", false, "Print JSON instead of text")
	return cmd
}

var showTemplateInfo = Info{
	Command:   "show-template",
	Aliases:   []string{"show", "sht"},
	Arg:       "[templateName]",
	DescShort: "Show template overview",
	DescLong:  "Displays a brief overview of a saved template: its windows and the named panes in each. Templates are stored in ~/.config/tmxu/templates/.",
	Examples: []string{
		"tmxu show-template default",
		"tmxu sht default",
	},
}

func (c *ShowTemplate) Flags() *flag.FlagSet { return c.fs }
func (c *ShowTemplate) Info() Info           { return showTemplateInfo }
func (c *ShowTemplate) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseSaveTemplateCmdFlags, err)
	}

	templateName := c.fs.Arg(0)
	if templateName == "" {
		return ErrSessionNameNotProvided
	}

	data, err := c.cfg.ShowTemplate(templateName)
	if err != nil {
		return err
	}

	snapshot, err := tmux.LoadSnapshot(data)
	if err != nil {
		return err
	}

	if c.json {
		out, err := snapshot.JSON()
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(c.ios.Out, string(out))
		return err

	}

	_, err = fmt.Fprintf(c.ios.Out, snapshot.String())
	return err
}
