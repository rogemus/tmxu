package main

import (
	_ "embed"
	"strings"

	"github.com/rogemus/tmxu/cli"
)

//go:embed version.txt
var version string

func main() {
	cli := cli.NewCli(strings.TrimSpace(version))
	cli.Run()
}
