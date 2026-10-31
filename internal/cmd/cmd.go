// Package cmd defines the tmxu commands. Each command implements the Cmd
// interface, which exposes its name, help text and a Run function.
package cmd

type Cmd interface {
	Command() string
	Aliases() []string
	Arg() string
	DescShort() string
	DescLong() string
	Flags() [][]string
	Examples() []string
	Run() error
}
