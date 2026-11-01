package main

import (
	_ "embed"
	"github.com/rogemus/tmxu/internal/cli"
	"github.com/rogemus/tmxu/internal/config"
)

//go:embed version.txt
var version string

func main() {
	// cli := cli.NewCli(strings.TrimSpace(version))
	// d, err := tmux.GetTmuxData()
	// if err != nil {
	// 	print("err:", err.Error(), "\n")
	// }

	cfg := config.New(version, "~/.config/tmxu")
	cli := cli.New(cfg)
	cli.Run()
	// cli.Run()
}
