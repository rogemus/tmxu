package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var ErrNotTerminal = errors.New("stdin is not a terminal; use -yes to skip the prompt")

// RequireTerminal returns ErrNotTerminal when in is a file that is not a
// terminal. Readers that are not files (for example in tests) are accepted.
func RequireTerminal(in io.Reader) error {
	f, ok := in.(*os.File)
	if !ok {
		return nil
	}

	if !term.IsTerminal(int(f.Fd())) {
		return ErrNotTerminal
	}

	return nil
}

func Confirm(prompt string, reader *bufio.Reader, out io.Writer) (bool, error) {
	fmt.Fprintf(out, "%s [y/N]: ", prompt)
	input, err := reader.ReadString('\n')

	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	answer := strings.ToLower(strings.TrimSpace(input))
	if errors.Is(err, io.EOF) && answer == "" {
		return false, errors.New("no answer on stdin; use -yes to skip the prompt")
	}

	switch answer {
	case "y", "yes":
		return true, nil
	}

	return false, nil
}
