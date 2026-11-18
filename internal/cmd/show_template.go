package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = (*ShowTemplate)(nil)

var (
	ErrCannotParseShowTemplateCmdFlags = fmt.Errorf("%w: cannot parse show template cmd flags", ErrUsage)
	ErrTemplateNameNotProvided         = fmt.Errorf("%w: template name not provided", ErrUsage)
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
	Arg:       "<templateName>",
	DescShort: "Show template overview",
	DescLong:  "Prints a saved template as a tree of its windows and the named panes in each. Templates are stored in ~/.config/tmxu/templates/. With -json prints the saved template JSON instead. Flags go before the template name.",
	Examples: []string{
		"tmxu show-template default",
		"tmxu show default",
		"tmxu sht default",
		"tmxu show-template -json default",
	},
}

func (c *ShowTemplate) Flags() *flag.FlagSet { return c.fs }
func (c *ShowTemplate) Info() Info           { return showTemplateInfo }
func (c *ShowTemplate) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseShowTemplateCmdFlags, err)
	}

	templateName := c.fs.Arg(0)
	if templateName == "" {
		return ErrTemplateNameNotProvided
	}

	data, err := c.cfg.ShowTemplate(templateName)
	if err != nil {
		return err
	}

	snapshot, err := tmux.LoadSnapshot(data)
	if err != nil {
		return err
	}

	snapshot, err = tmux.DesanitizeTemplSnapshot(snapshot, templateName)
	if err != nil {
		return err
	}

	return dumpSnapshot(c.json, snapshot, c.ios.Out)
}
