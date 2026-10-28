# TODO

Details and reasoning for every item live in [REVIEW.md](REVIEW.md).
Policy: stdlib-first. Only `golang.org/x/term` stays as a dependency. Write things by hand to learn Go.

## 1. Bugs (fix first)

- [ ] Fix `HasSession`: "session absent" must be `(false, nil)`, not an error
- [ ] Fix menu: negative `strings.Repeat`, ignored `MakeRaw`/`ReadByte` errors, empty `items`, EOF busy loop, Ctrl-C (byte 3)
- [ ] Reject template names containing `/`, `..`, empty or leading `.` (path traversal)
- [ ] Validate empty session name in `new-session`
- [ ] Bounds-check tmux output parsing (`parts[2]`, `Panes[0]`, `[""]` from empty output)
- [ ] Stop hard-coding `base-index` / `pane-base-index` as 1
- [ ] `restore -force` must not kill the session the user is currently in
- [ ] `confirm` on non-TTY stdin must fail loudly instead of silently aborting
- [ ] Template layouts are saved but not applied in `new-session -templ`: decide and fix

## 2. Errors

- [ ] Lowercase error strings, no trailing `\n` or punctuation
- [ ] Wrap with `%w` everywhere; stop replacing causes with generic text
- [ ] Fix typos: "sassion", "abborded", "availabe", "Widnows", `"%s ]n"`
- [ ] Rename sentinels `errorSessionExists` -> `errSessionExists`, `errorAborded` -> `errAborted`
- [ ] Add sentinels `ErrNoSession`, `ErrNoTemplate`, `ErrNoTmux` (`exec.ErrNotFound`)
- [ ] Add typed `*TmuxError{Args, Stderr, Err}` from one `run()` helper
- [ ] Use `errors.Is(err, fs.ErrNotExist)` instead of `os.Stat` + `os.IsNotExist`
- [ ] Handle every swallowed error (`tmux.go:29,32,141,142,150`, `cli.go:114`, `config.go:109,123,210,226`)
- [ ] `Run(ctx, args) error`; single `os.Exit` in `main.go`
- [ ] Errors to stderr, data to stdout, `tmxu:` prefix on errors
- [ ] Exit codes: 0 ok, 1 failure, 2 usage, 130 interrupt
- [ ] Use `errors.Join` to collect restore errors

## 3. Architecture and structure

- [ ] Remove `os.Args` from commands: `Run func(args []string) error`
- [ ] Remove package-level `version` global and package-level `Cmd` vars
- [ ] Drop `t` prefixes: `Session`, `Window`, `Pane`, `Template`; move `tSessionSimple` next to menu code
- [ ] Use `int` for `Order` (not `int16`)
- [ ] Parse tmux output once into typed structs (`SessionInfo` with `time.Time`); stop passing tab-joined strings
- [ ] Delete `math.go`; use builtin `min`/`max`
- [ ] Replace `renderTable` with `text/tabwriter`
- [ ] Split `cmd.go` (553 lines) into per-command files
- [ ] Unexport or document everything exported; add package doc comment
- [ ] Decide: keep flat `cli/` or move to `internal/` packages

## 4. Duplication

- [ ] One `snapshotSession` for `save-sessions` and `save-template`
- [ ] One `restoreSession` for `restore-sessions` and `new-session -templ`
- [ ] One mapping of `errSessionExists` to a user message
- [ ] Generic `readJSON[T]` / `writeJSON` plus one `configDir()` helper
- [ ] Replace `hasX`/`createX` pairs with `os.MkdirAll`; use `filepath.Join` for all paths
- [ ] Print real resolved paths in messages instead of literal `~/.config/tmxu`
- [ ] Single `TEMP_VALUE` const, or better: drop `SessionName`/`SessionWindow`/`Path` from `Template`
- [ ] Drop `SessionName`/`SessionWindow` from persisted `Window`/`Pane` (compute from parent); add schema migration
- [ ] Target new tmux only (3.8+, JSON layouts); drop support for the old checksum layout format (`d2b3,198x53,0,0,1`)
- [ ] `layout` must be a real nested JSON object (`json.RawMessage`), not a double-encoded string; validate with `json.Valid` on save and fail with a clear error if tmux returns non-JSON
- [ ] Omit `layout` for single-pane windows (`omitempty`); window size comes from the terminal, not the saved layout. Save and restore it only when a window has more than one pane.
- [ ] Do not call `select-layout` on restore for windows without a layout
- [ ] Document the minimum tmux version in README and check it at startup (`tmux -V`)
- [ ] Define flags once per command; derive help from `FlagSet` (`VisitAll`); print or delete `Examples`
- [ ] Fix typos in examples ("tmux" instead of "tmxu")

## 5. Testing

- [ ] Add `Runner` interface (`Run`, `Attach`) + `execRunner`; `Tmux` struct holds it
- [ ] Add `fakeRunner` recording args and returning canned output
- [ ] Add config `store{dir}` so tests use `t.TempDir()`
- [ ] Table tests: `TimeSince`, `renderTable`
- [ ] Table tests: tmux arg building (`NewSession`, `NewWindow`, `startDirArgs`), `ListSessions` ordering, empty output
- [ ] Config tests: round trip, missing file, bad JSON, traversal name, non-`.json` file, delete missing
- [ ] Golden files in `testdata/` with `-update` flag (session JSON, help output, table)
- [ ] Command-level tests with injected stdin/stdout/stderr
- [ ] Optional integration test with private socket `tmux -L tmxu-test-<name>`, skipped if tmux missing
- [ ] Run tests with `-race`
- [ ] Use `t.Helper`, `t.Cleanup`, `t.Setenv` where appropriate

## 6. Context and concurrency

- [ ] Thread `ctx` through everything; `signal.NotifyContext` in `main`
- [ ] `exec.CommandContext` with 5s timeout for non-interactive calls; no timeout for `attach`
- [ ] Optional: parallel snapshot with `sync.WaitGroup`, indexed result slice and a channel semaphore (measure first)
- [ ] Keep restore sequential (server start race, index dependencies)
- [ ] Decide on Ctrl-C during restore: clean up or report the partial session

## 7. Unix-style CLI

- [ ] `-h`/`--help` globally and per command (exit 0, stdout); usage errors exit 2 on stderr
- [ ] `-V`/`--version`
- [ ] Standard help layout: Usage, Commands, Options, Environment, Exit status; print Examples
- [ ] Support `--` terminator and flags after positional args (today `-path` after a name is silently ignored)
- [ ] Short and long flag pairs via a small helper
- [ ] Missing required argument prints usage and exits 2
- [ ] Unknown command: "did you mean"
- [ ] `-q`/`-v` flags with `log/slog` to stderr; `-v` logs tmux args
- [ ] TTY detection (`x/term`); `NO_COLOR`; `--color=auto|always|never` if color is added
- [ ] Menu: TTY check with plain-list fallback; draw on stderr or `/dev/tty`; alternate screen
- [ ] Restore terminal on SIGTERM/SIGHUP via context cancellation
- [ ] `ls` / `list-templates`: plain output when piped, `--json`, `-0`
- [ ] `save -o -`, `restore -f FILE`, `show-template NAME`, `attach -` (name from stdin)
- [ ] `-y`/`--yes` to skip prompts
- [ ] `-n`/`--dry-run` on restore, new, delete-template, save
- [ ] `$TMUX` awareness: `switch-client` inside tmux, `attach` outside
- [ ] Exact-match targets `-t =name`
- [ ] `exec.LookPath("tmux")` at startup with a clear message
- [ ] "no server running" handled like `tmux ls`
- [ ] Roll back half-created sessions when restore fails
- [ ] `delete-template`: clear missing-file error and `-f`
- [ ] Explicit XDG paths (`$XDG_CONFIG_HOME`, else `~/.config`); keep macOS on `~/.config` (avoid `os.UserConfigDir`)
- [ ] `TMXU_CONFIG_DIR` env and `--config-dir` flag; `tmxu paths` command
- [ ] Config files `0600`, dirs `0700`; `writeFileAtomic` (temp file, `Sync`, `Rename`); keep a `.bak`
- [ ] `"version": 1` field in the saved JSON
- [ ] Multibyte-safe padding in table and menu
- [ ] `tmxu completion zsh|bash|fish` (hand-written, `//go:embed`); hidden `__complete` for session/template names
- [ ] Man page `man/tmxu.1` (hand-written roff)
- [ ] Handle SIGPIPE/EPIPE on closed stdout
- [ ] `//go:build !windows` or a clear unsupported-OS message
- [ ] Document the minimum tested tmux version

## 8. Repo hygiene and release

- [ ] Add `ci.yaml`: `go build`, `go vet`, gofmt diff check, `go test -race ./...`, staticcheck, `go mod tidy` diff
- [ ] Release on `v*` tag push; use git tag as the single version source (`debug.ReadBuildInfo` + `-ldflags`); delete `version.txt`
- [ ] Release workflow: `go build` matrix (darwin/linux x amd64/arm64), `sha256sum`, GitHub Release
- [ ] Pin GitHub Actions by SHA; add Dependabot
- [ ] Fix `justfile`: `run` must pass `{{ARGS}}`, consistent indentation, add `build`, `install`, `check`, groups; fix `lint`
- [ ] Anchor `.gitignore` (`/tmxu`), add `.DS_Store`, `dist/`, `.claude/`, `*.test`, `coverage.out`
- [ ] Fix README (Go "1.16+" is wrong; document tmux version, exit codes, completions, JSON format)
- [ ] Fix stale CLAUDE.md (`gh.go`, `Formula/`) or leave it untracked
- [ ] Add CONTRIBUTING.md, CHANGELOG.md, `.editorconfig`
- [ ] Homebrew tap or formula, with man page and completions (or remove references)

## Suggested learning order

1. Errors and `%w`
2. `Run(args) error` and exit codes
3. Runner interface and fake
4. Table tests
5. Config store and atomic writes
6. Dedup
7. `context` and signals
8. Goroutines
9. Flags, completion, man page
10. CI and release
