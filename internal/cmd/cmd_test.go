package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rogemus/tmxu/internal/config"
)

func testIOs(in string) (IOs, *bytes.Buffer, *bytes.Buffer) {
	var out, errOut bytes.Buffer

	return IOs{In: strings.NewReader(in), Out: &out, Err: &errOut}, &out, &errOut
}

func testConfig() config.Config {
	return config.Config{DirPath: "/test/config", Version: "1.2.3"}
}

func allCmds(ios IOs) []Cmd {
	cfg := testConfig()

	return []Cmd{
		NewAttachSession(cfg, ios),
		NewDump(cfg, ios),
		NewListSessions(cfg, ios),
		NewNewSession(cfg, ios),
		NewSaveSessions(cfg, ios),
		NewVersion(cfg, ios),
	}
}

func TestCmdInfo(t *testing.T) {
	ios, _, _ := testIOs("")
	seen := map[string]string{}

	for _, c := range allCmds(ios) {
		info := c.Info()

		t.Run(info.Command, func(t *testing.T) {
			if info.Command == "" {
				t.Fatal("Command is empty")
			}
			if info.DescShort == "" {
				t.Error("DescShort is empty")
			}
			if len(info.Examples) == 0 {
				t.Error("no examples")
			}

			// The help template reads index 0 (name) and 1 (description).
			for _, f := range info.Flags {
				if len(f) != 2 {
					t.Errorf("flag %q has %d fields, want 2 (name, description)", f, len(f))
				}
			}

			for _, name := range append([]string{info.Command}, info.Aliases...) {
				if name == "" {
					t.Error("empty command or alias name")
				}
				if other, dup := seen[name]; dup {
					t.Errorf("%q is used by %q and %q", name, other, info.Command)
				}
				seen[name] = info.Command
			}
		})
	}
}
