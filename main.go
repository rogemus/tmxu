package main

import (
	_ "embed"
	// "strings"
	//
	// "github.com/rogemus/tmxu/cli"
	"github.com/rogemus/tmxu/internal/tmux"
)

//go:embed version.txt
var version string

func main() {
	// cli := cli.NewCli(strings.TrimSpace(version))
	d, err := tmux.GetTmuxData()
	if err != nil {
		print("err:", err.Error(), "\n")
	}
	print(d.String())
	// cli.Run()
}
