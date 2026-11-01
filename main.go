package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/rogemus/tmxu/internal/cli"
	"github.com/rogemus/tmxu/internal/cmd"
	"github.com/rogemus/tmxu/internal/config"
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
		fmt.Fprintf(os.Stderr, "tmxu: %v\n", err)
		if errors.Is(err, cmd.ErrUsage) {
			os.Exit(2)
		}

		os.Exit(1)
	}
}
