package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

var errorAborded = errors.New("abborded")

type menuItem struct {
	name string
}

func newMenuItem(name string) menuItem {
	return menuItem{
		name: name,
	}
}

func interactiveMenu(items []menuItem) (menuItem, error) {
	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	reader := bufio.NewReader(os.Stdin)
	selected := 0

	for {
		fmt.Print("\033[H\033[2J")
		fmt.Print("Use ↑/↓ to navigate, Enter to select, q to quit\r\n")

		for i, item := range items {
			markSelected := " "

			if i == selected {
				markSelected = "x"
			}

			fmt.Printf("%s [%s] %s \r\n", "", markSelected, item.name)
		}

		b, _ := reader.ReadByte()
		switch b {
		case 27:
			if reader.Buffered() > 0 {
				seq := make([]byte, 2)
				reader.Read(seq)

				if seq[0] == '[' {
					switch seq[1] {
					case 'A': // arrow up
						selected -= 1
					case 'B': // arrow down
						selected += 1
					}
				}
			} else {
				return menuItem{}, errorAborded
			}
		case 13: // Enter
			fmt.Print("\033[H\033[2J")
			return items[selected], nil
		case 'q':
			fmt.Print("\033[H\033[2J")
			return menuItem{}, errorAborded
		}
	}
}
