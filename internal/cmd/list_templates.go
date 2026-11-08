package cmd

import (
	"flag"
	"fmt"
	"text/tabwriter"

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
	templates, err := c.cfg.ListTemplates()
	if err != nil {
		return err
	}

	if len(templates) == 0 {
		_, err = fmt.Fprintf(c.ios.Out, "no templates\n")
		return err
	}

	w := tabwriter.NewWriter(c.ios.Out, 2, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Index\tTemplate Name")
	for index, name := range templates {
		fmt.Fprintf(w, "%d\t%s\n", index+1, name)
	}

	return w.Flush()
}
