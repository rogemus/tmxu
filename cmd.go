package main

import "fmt"

type cmd struct {
	command   string
	alias     string
	arg       string
	descShort string
	run       func()
}

func (c cmd) helpShort() {
	fmt.Printf(" %10s %8s    %s \n", c.command, c.arg, c.descShort)
}
