package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var errorAborded = errors.New("abborded")

type menuItem interface {
	Desc() string
	Title() string
}

func interactiveMenu(items []menuItem) (menuItem, error) {
	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	itemCount := len(items)
	reader := bufio.NewReader(os.Stdin)
	selected := 0

	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	for {
		fmt.Printf("\033[H\033[2J")
		fmt.Printf(" tmux sessions \r\n")
		fmt.Printf(" ──────────────────────────────────────────────── \r\n\n")

		for i, item := range items {
			markSelected := " "

			if i == selected {
				markSelected = ">"
			}

			title := item.Title() + " " + strings.Repeat("·", 25-len(item.Title()))
			fmt.Printf("  %s %s %s \r\n", markSelected, title, item.Desc())
		}

		fmt.Printf("\n ──────────────────────────────────────────────── \r")
		fmt.Print("\n Use ↑/↓ to navigate, Enter to select, q to quit\r")

		b, _ := reader.ReadByte()
		switch b {
		case 27:
			if reader.Buffered() > 0 {
				seq := make([]byte, 2)
				reader.Read(seq)

				if seq[0] == '[' {
					switch seq[1] {
					case 'A': // arrow up
						selected = Max(selected-1, 0)
					case 'B': // arrow down
						selected = Min(selected+1, itemCount-1)
					}
				}
			} else {
				return nil, errorAborded
			}
		case 13: // Enter
			fmt.Print("\033[H\033[2J")
			return items[selected], nil
		case 'q':
			fmt.Print("\033[H\033[2J")
			return nil, errorAborded
		}
	}
}

func sessionsToMenuItems(sessions []string) []menuItem {
	var items []menuItem

	for _, s := range sessions {
		items = append(items, newTSessionSimple(s))
	}

	return items
}
