package utils

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func Confirm(prompt string, in io.Reader, err io.Writer) bool {
	fmt.Fprintf(err, "%s [y/N]: ", prompt)
	input, _ := bufio.NewReader(in).ReadString('\n')

	switch strings.ToLower(strings.TrimSpace(input)) {
	case "y", "yes":
		return true
	}
	return false
}
