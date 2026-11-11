package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = (*DeleteTemplate)(nil)

var (
	ErrCannotParseDeleteTemplateCmdFlags = fmt.Errorf("%w: cannot parse delete template cmd flags", ErrUsage)
)

type DeleteTemplate struct {
	ios IOs
	cfg config.Config

	fs *flag.FlagSet
}

func NewDeleteTemplate(ios IOs, cfg config.Config) *DeleteTemplate {
	cmd := &DeleteTemplate{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("delete-template", flag.ContinueOnError),
	}
	cmd.fs.SetOutput(io.Discard)

	return cmd
}

var deleteTemplateInfo = Info{
	Command:   "delete-template",
	Aliases:   []string{"dt"},
	Arg:       "[templateName]",
	DescShort: "Delete saved template",
	DescLong:  "Removes a template file from ~/.config/tmxu/templates/.",
	Examples: []string{
		"tmxu delete-template templateName",
		"tmxu dt templateName",
	},
}

func (c *DeleteTemplate) Flags() *flag.FlagSet { return c.fs }
func (c *DeleteTemplate) Info() Info           { return deleteTemplateInfo }
func (c *DeleteTemplate) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseSaveTemplateCmdFlags, err)
	}

	templateName := c.fs.Arg(0)
	if templateName == "" {
		return ErrSessionNameNotProvided
	}

	path, err := c.cfg.DeleteTemplate(templateName)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.ios.Out, "Template deleted: `%s`\n", path)
	return err
}
