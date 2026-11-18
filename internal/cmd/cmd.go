// Package cmd defines the tmxu commands. Each command implements the Cmd
// interface, which exposes its name, help text and a Run function.
package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

// ErrUsage marks errors caused by wrong invocation (unknown command, bad flag,
// missing argument). main maps it to exit code 2.
var ErrUsage = errors.New("usage error")

var (
	ErrSessionNameNotProvided  = fmt.Errorf("%w: session name not provided", ErrUsage)
	ErrTemplateNameNotProvided = fmt.Errorf("%w: template name not provided", ErrUsage)
)

type Info struct {
	Command   string
	Aliases   []string
	Arg       string
	DescShort string
	DescLong  string
	Examples  []string
}

// Usage returns the one-line synopsis of the command.
func (i Info) Usage() string {
	if i.Arg == "" {
		return "tmxu " + i.Command
	}

	return "tmxu " + i.Command + " " + i.Arg
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
	Flags() *flag.FlagSet
	Run(args []string) error
}
