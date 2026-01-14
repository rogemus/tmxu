package main

import (
	"fmt"
	"os"
)

func LogInfo(msg string, args ...any) {
	if len(args) > 0 {
		fmt.Printf(fmt.Sprintf("%s \n", msg), args)
	} else {
		fmt.Printf("%s \n", msg)
	}
}

func LogError(msg string, args ...any) {
	if len(args) > 0 {
		fmt.Printf(fmt.Sprintf("ERROR: %s \n", msg), args)
	} else {
		fmt.Printf("%s \n", msg)
	}

	os.Exit(1)
}
