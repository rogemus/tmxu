package main

import (
	_ "embed"
	"fmt"

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

	cfg, err := config.New(version)
	if err != nil {
		panic(fmt.Errorf("%w: cannot read config dir", err))
	}

	cli := cli.New(cfg)
	cli.Run()
	// cli.Run()
}
