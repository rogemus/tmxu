// Command tmxu saves and restores tmux sessions (windows, panes and layouts)
// and creates new sessions from templates.
package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rogemus/tmxu/internal/cli"
	"github.com/rogemus/tmxu/internal/cmd"
	"github.com/rogemus/tmxu/internal/config"
	"github.com/rogemus/tmxu/internal/tmux"
	"github.com/rogemus/tmxu/internal/utils"
)

//go:embed version.txt
var version string

func main() {
	ios := cmd.IOs{
		Out: os.Stdout,
		In:  os.Stdin,
		Err: os.Stderr,
	}

	cfg, err := config.New(version)
	if err != nil {
		fmt.Fprintf(ios.Err, "tmxu: %v\n", err)
		os.Exit(1)
	}

	c := cli.New(cfg, ios)
	err = c.Run(os.Args[1:])
	if err != nil {
		printError(ios.Err, err)
		if errors.Is(err, cmd.ErrUsage) {
			os.Exit(2)
		}

		os.Exit(1)
	}
}

// printError writes err to w. Errors that have a next step get a hint line.
func printError(w io.Writer, err error) {
	var exists *tmux.SessionExistsError
	if errors.As(err, &exists) {
		fmt.Fprintf(w, "tmxu: session %s already exists\n", utils.Quote(exists.Name))
		fmt.Fprintf(w, "Run `tmxu attach %s` to use it\n", utils.Quote(exists.Name))
		return
	}

	fmt.Fprintf(w, "tmxu: %v\n", err)
}
