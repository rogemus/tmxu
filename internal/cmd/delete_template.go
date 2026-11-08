package cmd

import (
	"flag"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = (*DeleteTemplate)(nil)

type DeleteTemplate struct {
	ios IOs
	cfg config.Config
}

func NewDeleteTemplate(ios IOs, cfg config.Config) *DeleteTemplate {
	return &DeleteTemplate{ios: ios, cfg: cfg}
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

func (c *DeleteTemplate) Flags() *flag.FlagSet { return nil }
func (c *DeleteTemplate) Info() Info           { return deleteTemplateInfo }
func (c *DeleteTemplate) Run(args []string) error {
	return nil
}
