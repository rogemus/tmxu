# TODO

Everything left to do to finish moving the old `cli/` package into `internal/` and to ship `tmxu` as a good Unix tool. Reasoning for every item lives in [AI-REVIEW.md](AI-REVIEW.md).

Policy: stdlib-first. Only `golang.org/x/term` stays as a dependency. Write things by hand to learn Go.
Scope: single user, so there is no migration from the old `tmux-sessions.json`. The new schema is designed clean.
Target: tmux 3.8+ only (JSON `window_layout`).

Status: `main` is wired to `internal/cli`. `Run` dispatches `ls`, `save`, `version`, `help`, `attach`, `new-session`. Restore, templates and the menu are not ported. `save` writes nothing yet. Unit tests cover `cli` (100%), `utils` (100%), `cmd` (about 69%), `config` and `tmux` (about 60%); everything that shells out to tmux needs the runner seam.

## 1. Restore safety (fix first)

- [ ] `HasSession` returns `(false, nil)` for an absent session (exit code 1 from `has-session`); typed error for real failures
- [ ] Check `has-session` before creating; skip or refuse existing sessions (today only `NewSession` detects a duplicate, from tmux's stderr; a duplicate in a `CreateSessions` chain stops the rest)
- [ ] Never use `-k` on a session that was not created in this run
- [ ] Explicit force path with a prompt that lists what will be killed; refuse to kill the current session (`$TMUX`, `display-message -p '#S'`)
- [ ] Exact-match targets everywhere (`=name`, `'=name:idx'`); decide what to do with names containing `:` or `.` (target by session id, or reject)
- [ ] Read the real `base-index` (use or delete `GetBaseIndex`); handle saved index != current base so no blank leftover window stays
- [ ] Roll back sessions created in this run when a later step fails
- [ ] Ctrl-C during restore: clean up or report the partial session

## 2. Restore parity with the old `cli/`

- [ ] Window layout: read `#{window_layout}`, store it, apply `select-layout` after the splits
- [ ] Layout only for windows with more than one pane (`omitempty`); no `select-layout` for single-pane windows
- [ ] Decide split direction and the 80x24 detached-size problem (`new-session -x -y`, or apply layout per split)
- [ ] Start directory guard (old `startDirArgs`): verified that tmux does not fail on a missing `-c` directory, it silently starts in `$HOME`; decide whether to warn
- [x] `new-session` without `-c` starts in the client's current directory (verified), so no `os.Getwd` default is needed
- [ ] Pane naming rule: derived name, `@name` pane option, or raw title; escape `#` if `-T` expands formats
- [ ] Set pane titles without changing the active pane
- [ ] Restore the original active pane and window
- [ ] Preserve session order (`session_created` as `time.Time`; slice, not map; numeric sort of `$N`)
- [ ] `attach` / `switch-client` when `$TMUX` is set; `=name` exact target

## 3. Parsing and data model

- [ ] Stop building JSON inside the tmux format string (`GetTmuxData`, `ListSessions`): a `"`, `\` or newline in a name or title fails every row; use a safe delimiter or escaping
- [ ] Add `session_created` and `window_layout` to the parsed line
- [ ] Split DTO (tmux line) from domain and persisted types; consistent key names
- [ ] Persisted shape drops runtime ids (`$3`, `@7`, `%9`), `Window.SessionName`, `order` and the redundant `pane.name`; targets are computed at restore time
- [ ] `layout` stored as a real nested object (`json.RawMessage`), validated with `json.Valid` on save
- [ ] `"version"` field in the saved file
- [ ] Templates are a `Session` without runtime fields; delete `TEMP_VALUE`
- [ ] Return errors instead of `panic` for a window without panes; validate loaded data before building commands
- [ ] Exported accessors for `Data` (or return plain slices)
- [ ] Single map lookup in `parseTmuxLines`; drop `make(map, 0)` hints

## 4. Errors and execution

- [ ] One `run()` helper or a `Runner` interface (`Run`, `Attach`) with an `execRunner`; no direct `exec.Command` elsewhere
- [ ] Pass `[]string` argv with `";"` elements, or run a script through `source-file -`, instead of one joined string
- [ ] Typed `*TmuxError{Args, Stderr, Err}`; include stderr in messages
- [ ] Sentinels `ErrNoSession`, `ErrNoTemplate`, `ErrNoTmux` (`exec.ErrNotFound`); "tmux not found in PATH" message
- [ ] "No server running" handled like `tmux ls`
- [ ] `exec.LookPath("tmux")` and `tmux -V` check at startup with a clear message
- [ ] Wrap the cause in `GetBaseIndex` with `%w`
- [x] Fix `ErrCannotCreateSessions` text (now `ErrCannotCreateSession`)
- [x] `NewSession` reports tmux's own message and returns `ErrSessionExists` for a duplicate (`newSessionError`)
- [ ] `errors.Is(err, fs.ErrNotExist)` instead of `os.Stat` + `os.IsNotExist`
- [ ] `errors.Join` to collect restore errors
- [ ] Thread `ctx`; `signal.NotifyContext` in `main`; `exec.CommandContext` with a 5s timeout for non-interactive calls, none for `attach`
- [ ] Errors to stderr, data to stdout (already true for `ls` and `version`; hold the line for new commands)
- [ ] Exit code 130 on interrupt (0, 1, 2 done)
- [ ] Optional: parallel snapshot with `sync.WaitGroup`, indexed result slice and a semaphore (measure first); restore stays sequential

## 5. Port the remaining commands into `internal/`

- [ ] `save`: actually write the sessions file (one shared `snapshotSession` for sessions and templates)
- [ ] `restore-sessions` (one shared `restoreSession` for restore and new-from-template)
- [x] `new-session`: empty name and extra arguments rejected (`ErrSessionNameNotProvided`, `ErrUnexpectedArgument`), `-path` made absolute, duplicate returns `ErrSessionExists`
- [ ] `new-session -templ` fails with "not supported yet" until templates are ported (today it is parsed and silently ignored)
- [ ] `new-session`: names with `.` or `:` (tmux may rewrite them, so the printed name and `attach` could disagree; verify, then reject or print the real name)
- [ ] Map `ErrSessionExists` to a user message in one place
- [ ] `save-template`, `list-templates`, `delete-template` (clear missing-file error, `-f`)
- [ ] Template names validated: reject empty, `/`, `\`, `..`, leading `.`
- [ ] Template list skips non-`.json` files and subdirectories; a bad file is reported without failing the list
- [ ] `attach -menu`: implement or remove the flag and its example; fix the `tmux attach-session -menu` typo; drop the blank first line of `Run`
- [ ] Menu: `term.IsTerminal` check with a plain-list fallback; handle `MakeRaw` and read errors, EOF, Ctrl-C (byte 3), empty items, titles longer than the pad width; multibyte-safe padding
- [ ] Menu: draw on stderr or `/dev/tty`; alternate screen; escapes as constants; restore the terminal on SIGTERM/SIGHUP through context cancellation
- [ ] `utils.Confirm`: one reader, handle the read error, fail loudly when stdin is not a TTY, rename the `err io.Writer` param to `out`
- [ ] `AttachToSession` uses the injected `cmd.IOs` instead of `os.*`; fix the `%w %s: %w` error format
- [ ] Commands take their tmux dependency through an interface instead of calling package functions
- [ ] Config store with files `0600`, dirs `0700`, `writeFileAtomic` (temp file, `Sync`, `Rename`), keep a `.bak`
- [ ] Config: create the dir on first use; `TMXU_CONFIG_DIR` env and `--config-dir`; ignore a relative `XDG_CONFIG_HOME`; doc comments on `Config`, `DirPath`, `New`
- [ ] Use `cfg.DirPath` in messages and descriptions instead of the literal `~/.config/tmxu`
- [ ] Generic `readJSON[T]` / `writeJSON`; `filepath.Join` for all paths; `os.MkdirAll`
- [ ] Delete the old top-level `cli/` package once parity is reached

## 6. Unix-style CLI surface

- [ ] `-V` / `--version` globally
- [ ] Help layout: Usage, Commands, Options, Environment, Exit status
- [ ] Flags after positional args work; `--` terminator honored (today `-path` after a name is silently ignored)
- [ ] Missing required argument prints usage and exits 2 (done for `attach` and `new-session`; apply to the rest)
- [ ] Unknown command: "did you mean"
- [ ] `Info.Flags` as `[]Flag{Name, Desc}` instead of `[][]string`; help derived from the `FlagSet` (`VisitAll`)
- [ ] Detect alias collisions in `newCmd` (a test now catches a collision among registered commands; `newCmd` still overwrites silently)
- [ ] `-y` / `--yes`, `-n` / `--dry-run` (restore, new, delete-template, save), `-q` / `-v` with `log/slog` to stderr (`-v` logs tmux args)
- [ ] `ls` / `list-templates`: plain output when piped, `--json`, `-0`
- [ ] `save -o -`, `restore -f FILE`, `show-template NAME`, `attach -` (name from stdin)
- [ ] `tmxu paths` prints the resolved directories
- [ ] TTY detection; `NO_COLOR` and `TERM=dumb` respected if color is added
- [ ] Handle SIGPIPE / EPIPE on closed stdout (`tmxu ls | head -1`)
- [ ] `//go:build !windows` or a clear unsupported-OS message
- [ ] Decide what bare `tmxu` does (help + exit 0 today)
- [ ] `wantsHelp` becomes a plain function; remove unused `cfg` fields in `AttachSession`, `ListSessions`, `Version`
- [ ] `Cli.help(w, args)` passes `c.ios.Out` to `helpAll` instead of `w`
- [ ] `attachSessionInfo` has no `Flags` for `-menu`, so `tmxu help attach` omits it; fix the `"tmux s"` example in `save-sessions` (`tmxu s`)
- [ ] Doc comments on exported identifiers; unexport what is not API; package doc for `internal/cli`
- [ ] Replace `Comperable`/`Eq` with `slices.ContainsFunc` and a `map[string]struct{}` set
- [ ] Rename `GetTmuxData` -> `Snapshot`, `GetBaseIndex` -> `BaseIndex`; drop `GetName`/`GetTitle`/`GetPath` (call `utils.Quote` at the call site, or `Quoted*`)
- [ ] Rename `baseIndex.go` to `base_index.go`

## 7. Testing

- [ ] `fakeRunner` recording args and returning canned output
- [ ] Tests for `CreateSessions`, `CreateWindows`, `Snapshot`, `ListSessions` through the fake
- [ ] Names with `'`, `"`, `\`, `$`, `#`, spaces, newline (quoting of pane and window values is covered; session names and the tmux output format are not)
- [ ] `parseTmuxLines` ordering and a duplicate-name case
- [x] `config.New` (`t.Setenv` for `XDG_CONFIG_HOME` and `HOME`, version trimming, missing home)
- [ ] Config store round trip, missing file, bad JSON, traversal name, non-`.json` file, delete missing, using `t.TempDir()` (store not written yet)
- [x] `internal/cli` help output (order, aliases), command lookup, `ErrUnknownCommand`, `-h` handling
- [x] `utils.Confirm`, `Line` conversions, pane builders, `buildNewSessionCmd`, `newSessionError`
- [ ] Command-level tests with injected stdin/stdout/stderr (done for `version`, usage errors of `new-session` and `attach`, declined `save`; the rest needs the runner)
- [ ] Add a `Run` test for each command's success path once a fake runner exists
- [ ] Golden files in `testdata/` with an `-update` flag (saved JSON, help output, `ls` table)
- [ ] Optional integration test on a private socket (`tmux -L tmxu-test-<name>`), skipped without tmux
- [ ] Verify on a throwaway server (private socket, `-L`): single-string chain parsing, `'it'\''s'` quoting, `new-window -k` on a session's only window, `-T` format expansion, order of `list-panes -a`, session names with `.` or `:`
  - Verified: duplicate `new-session` exits 1 with `duplicate session: NAME` and stops a chain; `new-session -A -d` on an existing session fails without a terminal; `-c` with a missing directory exits 0 and falls back to `$HOME`
- [ ] Run tests with `-race`; use `t.Helper`, `t.Cleanup`, `t.Setenv`

## 8. Ship it (after the migration)

Do these last. Completion and the man page are generated from or written against the command registry, so they only make sense once the CLI surface settles.

- [ ] `ci.yaml`: `go build`, `go vet`, gofmt diff check, `go test -race ./...`, staticcheck, `go mod tidy` diff
- [ ] Release on `v*` tag push; git tag is the single version source (`-ldflags -X`, `-trimpath`, `CGO_ENABLED=0`, fallback to `debug.ReadBuildInfo`); delete `version.txt` and the `version` command's GitHub check if it is no longer needed
- [ ] Release workflow: build matrix (darwin/linux x amd64/arm64), `sha256sum`, `gh release create`
- [ ] Pin GitHub Actions by SHA; add Dependabot for modules and actions
- [ ] `tmxu completion zsh|bash|fish` (hand-written, `//go:embed` or generated from the registry); hidden `__complete` for session and template names; tests for it
- [ ] Man page `man/tmxu.1` (NAME, SYNOPSIS, COMMANDS, OPTIONS, FILES, ENVIRONMENT, EXIT STATUS, EXAMPLES, SEE ALSO)
- [ ] README: fix Go "1.16+" (go.mod requires 1.26), document tmux 3.8+, install, completion setup, exit codes, JSON format, an fzf recipe, how it differs from tmux-resurrect, tmuxinator, sesh
- [ ] `justfile`: `run` passes `{{ARGS}}`, consistent indentation, add `build`, `install`, `check`, groups; fix `lint`
- [ ] `.gitignore`: anchor `/tmxu`; add `.DS_Store`, `dist/`, `.claude/`, `*.test`, `coverage.out`
- [ ] Fix stale `CLAUDE.md` (`gh.go`, `Formula/`, old `cli/` layout) or leave it untracked
- [ ] `CONTRIBUTING.md`, `CHANGELOG.md`, `.editorconfig`
- [ ] Homebrew tap or formula, installing the man page and completions (or remove references)

## 9. Optional: restore performance (parked)

Symptom: after restore, panes "catch up" (shells still booting, prompts redraw) when the user attaches. Baseline first: `time tmxu restore`. If under about 2 s, items 1 and 2 are probably enough.

- [ ] `new-session -x <cols> -y <rows>` from `term.GetSize`; no flags when stdout is not a TTY
- [ ] Wait for panes to be ready: poll `capture-pane -p -t <pane>` every ~50 ms, ~5 s timeout per session, sessions in parallel; `waiting for shells...` on stderr; warn on timeout; `-no-wait` to skip
- [ ] Batch a window's tmux calls into one invocation (low priority)
- [ ] Staggered start in batches of 3-4 sessions (only if still janky)

## 10. Decisions

- [ ] One batched chain, separate calls, or `source-file -`
- [ ] Pane names: derived, `@name`, or raw titles
- [ ] On-disk format: reuse the `internal/tmux` structs or a separate persisted shape (recommended: separate)
- [ ] Existing sessions on restore: skip, refuse, or force
- [ ] Panic or error for a window without panes (recommended: error)
- [ ] Wait for shells by default with `-no-wait`, or opt-in `-wait`

## Suggested order

1. Runner seam, `TmuxError`, typed parsing (sections 4, 3)
2. Restore safety and parity (sections 1, 2) with tests on the fake
3. Persisted shape, config store, atomic writes (sections 3, 5)
4. Port the commands, then delete `cli/` (section 5)
5. `context` and signals (section 4)
6. CLI surface (section 6)
7. Ship it (section 8)
