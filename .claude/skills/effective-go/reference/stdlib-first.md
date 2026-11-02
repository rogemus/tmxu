# Stdlib first

Rule: before adding a dependency, check whether the standard library already does it. Browse https://pkg.go.dev/std.

This repo allows only `golang.org/x/term`. Flag anything else.

| Need | Stdlib | Instead of |
|------|--------|-----------|
| CLI flags and subcommands | `flag` (`flag.NewFlagSet` per subcommand) | cobra, urfave/cli, pflag |
| Structured logs | `log/slog` | logrus, zap, zerolog |
| JSON | `encoding/json` (`json.RawMessage`, `Decoder`) | jsoniter, gjson |
| Running programs | `os/exec` (`CommandContext`, `LookPath`) | go-cmd, sh libs |
| Paths | `path/filepath`, `os.UserConfigDir`, `os.UserHomeDir` | homedir, xdg libs |
| Files | `os`, `io/fs`, `os.ReadFile`, `os.WriteFile`, `os.MkdirAll` | afero |
| Slice/map helpers | `slices`, `maps`, `cmp` | lodash-style libs, samber/lo |
| Sets | `map[T]struct{}` | golang-set |
| Errors | `errors`, `fmt.Errorf("%w")`, `errors.Join` | pkg/errors |
| Tests | `testing`, `t.TempDir`, `t.Setenv`, `t.Cleanup`, `testing/fstest`, `net/http/httptest` | testify, gomock |
| HTTP client/server | `net/http` | resty, gin, echo (for simple needs) |
| Version compare | `golang.org/x/mod/semver` is not stdlib; write a small parser first | semver packages |
| Tables | `text/tabwriter` | tablewriter |
| Templating | `text/template` | external template engines |
| Cancellation, signals | `context`, `os/signal` (`NotifyContext`) | custom plumbing |
| Concurrency groups | `sync.WaitGroup`, channels | errgroup (x/ repo, not core) |
| Embedding files | `embed` | go-bindata |
| Build info | `runtime/debug.ReadBuildInfo` | version libs |
| Terminal raw mode | `golang.org/x/term` (already allowed) | termbox, tcell for a small menu |

## When a dependency is fine

- The stdlib genuinely lacks it (e.g. raw terminal mode).
- The code is security-sensitive and well-audited externally (crypto beyond `crypto/*`).
- The user decides it consciously and writes it down in `TODO.md`.

## Review questions

- What does this package do that `pkg.go.dev/std` does not?
- How many lines would the stdlib version take? Would you learn more writing it?
- Does this import pull in a transitive tree you do not want in `go.sum`?
