// Command tmxu saves and restores tmux sessions (windows, panes and layouts)
// and creates new sessions from templates.
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/rogemus/tmxu/internal/cli"
	"github.com/rogemus/tmxu/internal/cmd"
	"github.com/rogemus/tmxu/internal/config"
)

var version = "dev"

func appVersion() string {
	if version != "dev" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return "dev"
}

func main() {
	version = appVersion()

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
		fmt.Fprintf(ios.Err, "tmxu: %v\n", err)
		if errors.Is(err, cmd.ErrUsage) {
			os.Exit(2)
		}

		os.Exit(1)
	}
}
