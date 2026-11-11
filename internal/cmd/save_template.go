package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
)

var _ Cmd = (*SaveTemplate)(nil)

var (
	ErrCannotParseSaveTemplateCmdFlags = fmt.Errorf("%w: cannot parse save template cmd flags", ErrUsage)
)

type SaveTemplate struct {
	ios IOs
	cfg config.Config

	fs   *flag.FlagSet
	name string
}

func NewSaveTemplate(ios IOs, cfg config.Config) *SaveTemplate {
	cmd := &SaveTemplate{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("save-template", flag.ContinueOnError),
	}
	cmd.fs.SetOutput(io.Discard)
	cmd.fs.StringVar(&cmd.name, "name", "", "Name of the template. Defaults to session name")
	return cmd
}

var saveTemplateInfo = Info{
	Command:   "save-template",
	Aliases:   []string{"st"},
	Arg:       "[sessionName]",
	DescShort: "Save session as template",
	DescLong:  "Saves a running tmux session as a reusable template. Templates are stored in ~/.config/tmxu/templates/.",
	Examples: []string{
		"tmxu st sessionName",
		"tmxu save-template sessionName",
		"tmxu save-template -name templateName sessionName",
	},
}

func (c *SaveTemplate) Flags() *flag.FlagSet { return c.fs }
func (c *SaveTemplate) Info() Info           { return saveTemplateInfo }
func (c *SaveTemplate) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseSaveTemplateCmdFlags, err)
	}

	sessionName := c.fs.Arg(0)
	if sessionName == "" {
		return ErrSessionNameNotProvided
	}

	hasSession, err := tmux.HasSession(sessionName)
	if err != nil || !hasSession {
		return err
	}

	snapshot, err := tmux.SessionSnapshot(sessionName)
	if err != nil {
		return err
	}

	templateName := sessionName
	if c.name != "" {
		templateName = c.name
	}

	snapshot = tmux.SanitizeSnapshot(snapshot, sessionName)
	out, err := snapshot.JSON()
	if err != nil {
		return err
	}

	path, err := c.cfg.SaveTemplate(out, templateName)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.ios.Out, "Templates saved at: %s\n", path)
	return nil
}
