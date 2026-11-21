package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = (*SaveTemplate)(nil)

var (
	ErrCannotParseSaveTemplateCmdFlags = fmt.Errorf("%w: cannot parse save template cmd flags", ErrUsage)
	ErrTemplateExists                  = errors.New("template already exists")
)

type SaveTemplate struct {
	ios IOs
	cfg config.Config

	fs    *flag.FlagSet
	name  string
	force bool
}

func NewSaveTemplate(ios IOs, cfg config.Config) *SaveTemplate {
	cmd := &SaveTemplate{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("save-template", flag.ContinueOnError),
	}
	cmd.fs.SetOutput(io.Discard)
	cmd.fs.StringVar(&cmd.name, "name", "", "Name of the template. Defaults to session name")
	cmd.fs.BoolVar(&cmd.force, "force", false, "override existing template while saving")
	return cmd
}

var saveTemplateInfo = Info{
	Command:   "save-template",
	Aliases:   []string{"st"},
	Arg:       "<sessionName>",
	DescShort: "Save session as template",
	DescLong:  "Saves a running tmux session as a reusable template in %s. The session name and every pane path are replaced with placeholders, so a template can be reused under any name. -name sets the template name; it defaults to the session name. An existing template with the same name is refused unless -force is given. Flags go before the session name.",
	Examples: []string{
		"tmxu st sessionName",
		"tmxu save-template sessionName",
		"tmxu save-template -name templateName sessionName",
		"tmxu save-template -force -name templateName sessionName",
	},
}

func (c *SaveTemplate) Flags() *flag.FlagSet { return c.fs }
func (c *SaveTemplate) Info() Info {
	info := saveTemplateInfo
	info.DescLong = fmt.Sprintf(info.DescLong, c.cfg.TemplatesDir())
	return info
}
func (c *SaveTemplate) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseSaveTemplateCmdFlags, err)
	}

	sessionName := c.fs.Arg(0)
	if sessionName == "" {
		return ErrSessionNameNotProvided
	}

	templateName := sessionName
	if c.name != "" {
		templateName = c.name
	}
	if strings.TrimSpace(templateName) == "" {
		return ErrTemplateNameNotProvided
	}

	snapshot, err := tmux.Snapshot(sessionName)
	if err != nil {
		return err
	}

	hasTemplate, err := c.cfg.HasTemplate(templateName)
	if err != nil {
		return err
	}

	if hasTemplate && !c.force {
		return fmt.Errorf("%w: %q (use -force to overwrite)", ErrTemplateExists, templateName)
	}

	snapshot = snapshot.AsTemplate()
	out, err := snapshot.JSON()
	if err != nil {
		return err
	}

	path, err := c.cfg.SaveTemplate(out, templateName)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.ios.Out, "Templates saved at: %s\n", path)
	return err
}
