package tmux

import (
	"errors"
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
		return 0, ErrCannotGetBaseIndex
	}

	i, err := strconv.Atoi(strings.TrimSpace(string(o)))
	if err != nil {
		return 0, ErrCannotGetBaseIndex
	}

	return i, nil
}
