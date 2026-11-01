// Package cmd defines the tmxu commands. Each command implements the Cmd
// interface, which exposes its name, help text and a Run function.
package cmd

import "io"

type Info struct {
	Command   string
	Aliases   []string
	Arg       string
	DescShort string
	DescLong  string
	Flags     [][]string
	Examples  []string
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
