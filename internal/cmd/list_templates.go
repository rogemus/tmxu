package cmd

import (
	"flag"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = (*ListTemplates)(nil)

type ListTemplates struct {
	ios IOs
	cfg config.Config
}

func NewListTemplates(ios IOs, cfg config.Config) *ListTemplates {
	return &ListTemplates{ios: ios, cfg: cfg}
}

var listTemplatesInfo = Info{
	Command:   "list-templates",
	Aliases:   []string{"lt"},
	DescShort: "List all saved templates",
	DescLong:  "Displays all saved templates with their windows and panes. Templates are stored in ~/.config/tmxu/templates/.",
	Examples: []string{
		"tmxu list-templates",
		"tmxu lt",
	},
}

func (c *ListTemplates) Flags() *flag.FlagSet { return nil }
func (c *ListTemplates) Info() Info           { return listTemplatesInfo }
func (c *ListTemplates) Run(args []string) error {
	return nil
}
