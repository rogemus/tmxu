package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/rogemus/tmxu/internal/tmux"
)

var (
	ErrCannotEncodeJSON = errors.New("cannot encode json")
)

func printJSON(w io.Writer, v any) error {
	out, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotEncodeJSON, err)
	}

	_, err = fmt.Fprintln(w, string(out))
	return err
}

func dumpSnapshot(asJSON bool, data tmux.SnapshotData, w io.Writer) error {
	if asJSON {
		return printJSON(w, data)
	}

	_, err := fmt.Fprintln(w, data.String())
	return err
}
