package cli

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/rogemus/tmxu/internal/cmd"
	"github.com/rogemus/tmxu/internal/config"
)

type stubCmd struct {
	info cmd.Info
	err  error

	runs *int
	args *[]string
}

func newStub(info cmd.Info, err error) stubCmd {
	return stubCmd{info: info, err: err, runs: new(int), args: new([]string)}
}

func (s stubCmd) Info() cmd.Info { return s.info }
func (s stubCmd) Run(args []string) error {
	*s.runs++
	*s.args = args

	return s.err
}

var (
	alphaInfo = cmd.Info{
		Command:   "alpha",
		Aliases:   []string{"al", "a"},
		Arg:       "[name]",
		DescShort: "First test command",
		DescLong:  "Longer text of the first test command.",
		Flags:     [][]string{{"force", "Do it anyway"}},
		Examples:  []string{"tmxu alpha one", "tmxu al one"},
	}
	betaInfo = cmd.Info{
		Command:   "beta",
		DescShort: "Second test command",
	}
)

func newTestCli(alphaErr error) (c *Cli, alpha, beta stubCmd, out, errOut *bytes.Buffer) {
	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	alpha, beta = newStub(alphaInfo, alphaErr), newStub(betaInfo, nil)

	c = &Cli{
		cmds: make(map[string]cmd.Cmd),
		ios:  cmd.IOs{In: strings.NewReader(""), Out: out, Err: errOut},
	}
	c.newCmd(alpha)
	c.newCmd(beta)

	return c, alpha, beta, out, errOut
}

func TestRunGeneralHelp(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"-h"}, {"--help"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			c, alpha, _, out, errOut := newTestCli(nil)

			if err := c.Run(args); err != nil {
				t.Fatalf("Run(%q) error = %v", args, err)
			}

			got := out.String()
			for _, want := range []string{"alpha (al, a)", "First test command", "beta", "tmxu help [command]"} {
				if !strings.Contains(got, want) {
					t.Errorf("help does not contain %q:\n%s", want, got)
				}
			}
			if errOut.Len() != 0 {
				t.Errorf("stderr = %q, want nothing", errOut)
			}
			if *alpha.runs != 0 {
				t.Error("help must not run a command")
			}
		})
	}
}

func TestRunHelpListsCommandsInRegistrationOrder(t *testing.T) {
	c, _, _, out, _ := newTestCli(nil)

	if err := c.Run([]string{"help"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := out.String()
	if a, b := strings.Index(got, "alpha"), strings.Index(got, "beta"); a < 0 || b < 0 || a > b {
		t.Errorf("alpha at %d, beta at %d, want alpha first:\n%s", a, b, got)
	}
}

func TestRunHelpForCommand(t *testing.T) {
	tests := [][]string{
		{"help", "alpha"},
		{"help", "al"},
		{"alpha", "-h"},
		{"alpha", "-help"},
		{"alpha", "--help"},
		{"al", "--help"},
		{"alpha", "one", "-h"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			c, alpha, beta, out, _ := newTestCli(nil)

			if err := c.Run(args); err != nil {
				t.Fatalf("Run(%q) error = %v", args, err)
			}

			got := out.String()
			for _, want := range []string{
				"alpha - First test command",
				"ALIASES:\n  al, a",
				"DESCRIPTION:\n  Longer text of the first test command.",
				"tmxu alpha [name] [flags]",
				"-force      Do it anyway",
				"EXAMPLES:\n  tmxu alpha one\n  tmxu al one",
			} {
				if !strings.Contains(got, want) {
					t.Errorf("help does not contain %q:\n%s", want, got)
				}
			}
			if *alpha.runs != 0 || *beta.runs != 0 {
				t.Error("help must not run a command")
			}
		})
	}
}

func TestRunHelpOmitsEmptySections(t *testing.T) {
	c, _, _, out, _ := newTestCli(nil)

	if err := c.Run([]string{"beta", "-h"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := out.String()
	for _, section := range []string{"ALIASES:", "DESCRIPTION:", "FLAGS:", "EXAMPLES:"} {
		if strings.Contains(got, section) {
			t.Errorf("help contains %s for a command without it:\n%s", section, got)
		}
	}
	if !strings.Contains(got, "USAGE:\n  tmxu beta\n") {
		t.Errorf("usage line is wrong:\n%s", got)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	tests := [][]string{
		{"nope"},
		{"help", "nope"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			c, alpha, _, out, _ := newTestCli(nil)

			err := c.Run(args)

			if !errors.Is(err, ErrUnknownCommand) {
				t.Fatalf("Run(%q) error = %v, want ErrUnknownCommand", args, err)
			}
			if !errors.Is(err, cmd.ErrUsage) {
				t.Errorf("error %v does not wrap ErrUsage, main would not exit 2", err)
			}
			if !strings.Contains(err.Error(), `"nope"`) {
				t.Errorf("error %q does not name the command", err)
			}
			if *alpha.runs != 0 {
				t.Error("unknown command ran a command")
			}
			if out.Len() != 0 {
				t.Errorf("stdout = %q, want nothing on an error", out)
			}
		})
	}
}

func TestRunUnknownCommandPrintsHelpToStderr(t *testing.T) {
	c, _, _, out, errOut := newTestCli(nil)

	_ = c.Run([]string{"nope"})

	if out.Len() != 0 {
		t.Errorf("stdout = %q, want nothing", out)
	}
	if !strings.Contains(errOut.String(), "alpha") {
		t.Errorf("stderr = %q, want the command list", errOut)
	}
}

func TestRunDispatch(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantArgs []string
	}{
		{name: "name", args: []string{"alpha"}, wantArgs: []string{}},
		{name: "alias", args: []string{"al"}, wantArgs: []string{}},
		{name: "second alias", args: []string{"a"}, wantArgs: []string{}},
		{name: "args are passed without the command name", args: []string{"alpha", "-force", "one"}, wantArgs: []string{"-force", "one"}},
		{name: "help after the terminator is an argument", args: []string{"alpha", "--", "-h"}, wantArgs: []string{"--", "-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, alpha, beta, out, _ := newTestCli(nil)

			if err := c.Run(tt.args); err != nil {
				t.Fatalf("Run(%q) error = %v", tt.args, err)
			}

			if *alpha.runs != 1 || *beta.runs != 0 {
				t.Errorf("runs = alpha %d, beta %d, want 1 and 0", *alpha.runs, *beta.runs)
			}
			if !reflect.DeepEqual(*alpha.args, tt.wantArgs) {
				t.Errorf("args = %q, want %q", *alpha.args, tt.wantArgs)
			}
			if out.Len() != 0 {
				t.Errorf("stdout = %q, want nothing", out)
			}
		})
	}
}

func TestRunReturnsCommandError(t *testing.T) {
	boom := errors.New("boom")
	c, _, _, _, _ := newTestCli(boom)

	if err := c.Run([]string{"alpha"}); !errors.Is(err, boom) {
		t.Errorf("Run() error = %v, want %v", err, boom)
	}
}

func TestWantsHelp(t *testing.T) {
	c := &Cli{}

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no args", args: nil, want: false},
		{name: "-h", args: []string{"-h"}, want: true},
		{name: "-help", args: []string{"-help"}, want: true},
		{name: "--help", args: []string{"--help"}, want: true},
		{name: "after other args", args: []string{"a", "b", "-h"}, want: true},
		{name: "terminator first", args: []string{"--", "-h"}, want: false},
		{name: "after the terminator only", args: []string{"a", "--", "--help"}, want: false},
		{name: "before the terminator", args: []string{"-h", "--"}, want: true},
		{name: "-help as part of a value", args: []string{"-path=-help"}, want: false},
		{name: "similar flag", args: []string{"-hh"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.wantsHelp(tt.args); got != tt.want {
				t.Errorf("wantsHelp(%q) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestNewRegistersCommandsWithoutCollisions(t *testing.T) {
	ios := cmd.IOs{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	c := New(config.Config{DirPath: "/test/config", Version: "1.2.3"}, ios)

	if len(c.cmdsOrder) == 0 {
		t.Fatal("New() registered no commands")
	}

	// A colliding alias overwrites an earlier entry, so the map ends up
	// smaller than the number of names handed out.
	names := 0
	for _, name := range c.cmdsOrder {
		names += 1 + len(c.cmds[name].Info().Aliases)
	}
	if len(c.cmds) != names {
		t.Errorf("%d names registered into %d map entries: an alias collides", names, len(c.cmds))
	}

	for _, name := range c.cmdsOrder {
		if got := c.cmds[name].Info().Command; got != name {
			t.Errorf("cmds[%q] is command %q", name, got)
		}
	}
}

func TestNewSupportsVersionEndToEnd(t *testing.T) {
	var out bytes.Buffer
	ios := cmd.IOs{In: strings.NewReader(""), Out: &out, Err: &bytes.Buffer{}}
	c := New(config.Config{Version: "1.2.3"}, ios)

	for _, name := range []string{"version", "v"} {
		out.Reset()

		if err := c.Run([]string{name}); err != nil {
			t.Fatalf("Run(%q) error = %v", name, err)
		}
		if want := "Current tmxu version 1.2.3\n"; out.String() != want {
			t.Errorf("Run(%q) stdout = %q, want %q", name, out.String(), want)
		}
	}
}
