// Package cmd defines the tmxu commands. Each command implements the Cmd
// interface, which exposes its name, help text and a Run function.
package cmd

import (
	"errors"
	"flag"
	"io"
)

// ErrUsage marks errors caused by wrong invocation (unknown command, bad flag,
// missing argument). main maps it to exit code 2.
var ErrUsage = errors.New("usage error")

type Info struct {
	Command   string
	Aliases   []string
	Arg       string
	DescShort string
	DescLong  string
	Flags     []Flag
	Examples  []string
	Flags2    *flag.FlagSet
}

type Flag struct {
	Name string
	Desc string
}

type IOs struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

type Cmd interface {
	Info() Info
	Run(args []string) error
}
