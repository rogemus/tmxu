package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
	"github.com/rogemus/tmxu/internal/utils"
)

var _ Cmd = (*RestoreSessions)(nil)

var (
	ErrCannotParseRestoreSessionsCmdFlags = fmt.Errorf("%w: cannot parse restore sessions cmd flags", ErrUsage)
	ErrSessionsAlreadyCreated             = errors.New("sessions already created")
)

type RestoreSessions struct {
	ios IOs
	cfg config.Config

	fs    *flag.FlagSet
	force bool
	yes   bool
}

func NewRestoreSessions(ios IOs, cfg config.Config) *RestoreSessions {
	cmd := &RestoreSessions{
		ios: ios,
		cfg: cfg,
		fs:  flag.NewFlagSet("restore-sessions", flag.ContinueOnError),
	}
	cmd.fs.SetOutput(io.Discard)
	cmd.fs.BoolVar(&cmd.force, "force", false, "override existing sessions while restoring")
	cmd.fs.BoolVar(&cmd.yes, "yes", false, "Skip the confirmation prompt")
	return cmd
}

var restoreSessionsInfo = Info{
	Command:   "restore-sessions",
	Aliases:   []string{"restore", "r"},
	DescShort: "Restore tmux sessions",
	DescLong:  "Recreates tmux sessions from %s. Refuses to run if any saved session already exists; -force kills those sessions first and recreates them (asks for confirmation unless -yes is given).",
	Examples: []string{
		"tmxu restore-sessions",
		"tmxu restore",
		"tmxu r",
		"tmxu restore -force",
	},
}

func (c *RestoreSessions) Flags() *flag.FlagSet { return c.fs }
func (c *RestoreSessions) Info() Info {
	info := restoreSessionsInfo
	info.DescLong = fmt.Sprintf(info.DescLong, c.cfg.SessionsFile())
	return info
}
func (c *RestoreSessions) Run(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotParseRestoreSessionsCmdFlags, err)
	}

	data, err := c.cfg.ReadSessionsFile()
	if err != nil {
		return err
	}

	snapshot, err := tmux.LoadSnapshot(data)
	if err != nil {
		return err
	}

	ls, err := tmux.ListSessions()
	if err != nil {
		return err
	}

	existing := snapshot.ExistingIn(ls)
	sesNames := make([]string, 0, len(existing))
	for _, es := range existing {
		sesNames = append(sesNames, es.Name)
	}
	existingNames := strings.Join(sesNames, ", ")
	hasExisting := len(existing) > 0

	if hasExisting && !c.force {
		return fmt.Errorf(
			"%w: existing sessions [%s] (use -force to override them)",
			ErrSessionsAlreadyCreated,
			existingNames,
		)
	}

	prompt := "Restore all tmux sessions?"
	if hasExisting {
		prompt = fmt.Sprintf("Restore all tmux sessions? Existing sessions [%s] will be killed.", existingNames)
	}

	if !c.yes {
		err := utils.RequireTerminal(c.ios.In)
		if err != nil {
			return err
		}

		ok, err := utils.Confirm(prompt, c.ios.Reader, c.ios.Err)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(c.ios.Err, "Aborted")
			return nil
		}
	}

	if hasExisting {
		err = tmux.KillSessions(existing)
		if err != nil {
			return err
		}
	}

	for _, s := range snapshot.Sessions {
		err = tmux.RestoreSession(s.Name, "", s.Windows)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(c.ios.Out, "Restored %d tmux sessions\n", len(snapshot.Sessions))
	return err
}
