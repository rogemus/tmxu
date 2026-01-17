package main

import (
	_ "embed"

	"github.com/rogemus/tmxu/cli"
)

//go:embed version.txt
var version string

func main() {
	cli := cli.NewCli(version)
	cli.Run()
}
