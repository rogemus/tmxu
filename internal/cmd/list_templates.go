package cmd

import (
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/rogemus/tmxu/internal/config"
)

var _ Cmd = (*ListTemplates)(nil)

var (
	ErrCannotParseListTemplatesCmdFlags = fmt.Errorf("%w: cannot parse list templates cmd flags", ErrUsage)
)

type ListTemplates struct {
	ios IOs
	cfg config.Config

	fs   *flag.FlagSet
	json bool
}

func NewListTemplates(ios IOs, cfg config.Config) *ListTemplates {
	cmd := &ListTemplates{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("list-templates", flag.ContinueOnError),
	}
	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.json, "json", false, "Print JSON instead of text")
	return cmd
}

var listTemplatesInfo = Info{
	Command:   "list-templates",
	Aliases:   []string{"lt"},
	DescShort: "List all saved templates",
	DescLong:  "Lists the names of all saved templates, numbered. Templates are stored in ~/.config/tmxu/templates/. With -json prints {\"templates\":[...]} on one line; no templates prints {\"templates\":[]}.",
	Examples: []string{
		"tmxu list-templates",
		"tmxu lt",
		"tmxu list-templates -json",
	},
}

func (c *ListTemplates) Flags() *flag.FlagSet { return c.fs }
func (c *ListTemplates) Info() Info           { return listTemplatesInfo }
func (c *ListTemplates) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseListTemplatesCmdFlags, err)
	}

	templates, err := c.cfg.ListTemplates()
	if err != nil {
		return err
	}

	if c.json {
		if templates == nil {
			templates = []string{}
		}

		return printJSON(c.ios.Out, map[string][]string{"templates": templates})
	}

	if len(templates) == 0 {
		_, err = fmt.Fprintln(c.ios.Err, "no templates")
		return err
	}

	w := tabwriter.NewWriter(c.ios.Out, 2, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Index\tTemplate Name")
	for index, name := range templates {
		fmt.Fprintf(w, "%d\t%s\n", index+1, name)
	}

	return w.Flush()
}
