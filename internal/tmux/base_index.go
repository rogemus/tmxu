package tmux

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var (
	ErrCannotGetBaseIndex = errors.New("cannot get base index")
)

func BaseIndex() (int, error) {
	o, err := exec.Command("tmux", "show-options", "-gv", "base-index").Output()
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrCannotGetBaseIndex, err)
	}

	i, err := strconv.Atoi(strings.TrimSpace(string(o)))
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrCannotGetBaseIndex, err)
	}

	return i, nil
}
