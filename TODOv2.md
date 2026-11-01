# TODO v2: new implementation (`internal/`)

Checklist for the rewrite in `internal/tmux`, `internal/utils`, `internal/cmd`, `internal/cli` and `internal/config`. Details and reasoning: [REVIEWv2.md](REVIEWv2.md). The old list for the top-level `cli/` package is [TODO.md](TODO.md).

Policy: stdlib-first. Write by hand to learn Go. `main` is wired to `internal/cli`, but `Run` only prints help so far; restore is not wired.

## 1. Restore safety (fix first)

- [ ] Check `has-session` before creating; skip or refuse existing sessions (today `-k` overwrites live windows)
- [ ] Never use `-k` on a session that was not created in this run
- [ ] Add an explicit force path with confirmation; refuse to kill the current session
- [ ] Exact-match targets: window and pane targets use `'=name:idx'` (confirmed in `man tmux` 3.8); still open: names containing `:` and `.`, or target by session id
- [ ] Read the real `base-index` (use or delete `GetBaseIndex`); handle saved index != current base (leftover blank window)
- [ ] Roll back sessions created in this run when a later step fails

## 2. Missing restore features (parity with the old `cli/`)

- [ ] Window layout: read `#{window_layout}`, store it, apply `select-layout` after the splits
- [ ] Decide split direction and the 80x24 detached-size problem (`new-session -x -y`, or apply layout per split)
- [ ] Start directory guard (old `startDirArgs`), or verify tmux's fallback for a missing directory
- [ ] Pane naming rule: derived name, `@name` pane option, or raw title (and escape `#` if tmux expands `-T`)
- [ ] Restore the original active pane and window
- [ ] Preserve session order (`session_created`, slice not map, numeric sort of `$N`); `ListSessions` now returns tmux's own order, not creation order
- [ ] `HasSession` returning `(false, nil)` for absent, typed
- [ ] `attach` / `switch-client` when `$TMUX` is set
- [ ] `restore-sessions`, `new-session`, templates (`TEMP_VALUE` or its replacement); `list-sessions` exists in `internal/cmd`
- [ ] Load and save the sessions file (`save-sessions` fetches data and discards it); templates directory

## 3. Parsing and data model

- [ ] Stop building JSON inside the tmux format in `GetTmuxData` and `ListSessions` (a `"`, `\` or newline in a title or name fails every row); use a safe delimiter or escaping
- [ ] Add `session_created` (`time.Time`) and `window_layout` to the parsed line
- [ ] Split DTO from domain: `Session` carries tmux's key names (`session_id`, `session_name`) while `Window` and `Pane` use `id` and `name`
- [ ] Decide what is persisted: drop runtime ids (`$3`, `@7`, `%9`) and `Window.SessionName` from the file
- [ ] Add a `"version"` field and a migration from the old `tmux-sessions.json`
- [ ] Return errors instead of `panic` for a window without panes; validate loaded data before building commands
- [ ] Exported accessors for `Data` (or return plain slices)

## 4. Errors and execution

- [ ] Pass `[]string` argv with `";"` elements, or run a script via `source-file -`, instead of one joined string
- [ ] Typed `*TmuxError{Args, Stderr, Err}` from one `run()` helper; include stderr in messages
- [ ] Wrap the cause in `GetBaseIndex` with `%w`
- [ ] Fix `ErrCannotCreateSessions` text (says "cannot create session")
- [ ] `ErrNoSession`, `ErrNoTemplate`, `ErrNoTmux` (`exec.ErrNotFound`); handle "no server running" like `tmux ls`
- [ ] Check `tmux -V` and `exec.LookPath("tmux")`
- [ ] Thread `ctx` and `exec.CommandContext` with a timeout (none for `attach`)

## 5. Idiomatic Go cleanup

- [ ] Replace `Comperable`/`Eq` with `slices.ContainsFunc` and a `map[string]struct{}` set
- [ ] Rename `GetTmuxData` and `GetBaseIndex` (stutter and `Get` prefix), e.g. `Snapshot()` and `BaseIndex()`
- [ ] Remove the `GetName`/`GetTitle`/`GetPath` quoting methods (non-idiomatic `Get` prefix, and they return a shell-quoted copy, not the field): call `utils.Quote(x.Name)` at the call site in the `build*Cmd` functions, or rename to `QuotedName`/`QuotedTitle`/`QuotedPath`; consider `Target()` for `Window.SessionIndex`
- [ ] Rename `baseIndex.go` to `base_index.go`
- [ ] Single map lookup in `parseTmuxLines`; drop `make(map, 0)` hints
- [ ] Doc comments on exported identifiers (the `cmd` types and constructors lost theirs); unexport what is not API
- [ ] Package doc comment for `internal/cli`

## 6. Testing

- [ ] `Runner` interface (`Run`, `Attach`) and `fakeRunner`; stop calling `exec.Command` directly
- [ ] Test `CreateSessions`, `CreateWindows`, `GetTmuxData`, `ListSessions` through the fake
- [ ] Tests for names with `'`, `"`, `\`, `$`, `#`, spaces, newline
- [ ] Test `parseTmuxLines` ordering and a duplicate-name case
- [ ] Tests for `config.New` (`t.Setenv` for `XDG_CONFIG_HOME` and `HOME`, version trimming)
- [ ] Tests for `cli` help output (order, aliases) and for command lookup
- [ ] Optional integration test on a private socket (`tmux -L tmxu-test-<name>`), skipped without tmux
- [ ] Verify on a throwaway server: single-string chain parsing, `'it'\''s'` quoting, `new-window -k` on a session's only window, `-c` with a missing directory, `-T` format expansion
- [ ] Run tests with `-race`

## 7. CLI surface

- [ ] `Cli.Run(args)`: dispatch on args, return an error, unknown-command handling, register `help`
- [ ] Wire `internal/tmux` into `restore`, `new-session`, `attach` (`ls` and `save` call it; `save` writes nothing)
- [ ] Help: range over `cmdsOrder` and drop the `seen` set (map order is random); use the `cmdName` argument; per-command help from `DescLong`, `Examples`, `Flags`; write to an injected `io.Writer`, with a trailing newline
- [ ] `Info.Flags` as `[]Flag{Name, Desc}` instead of `[][]string`
- [ ] Detect alias collisions in `newCmd`
- [ ] `ls`: drop the "Available sessions" header on a pipe (or send it to stderr); fix the trailing `\n ` in the format; use `tabwriter`
- [ ] `utils.Confirm`: one reader, handle the read error, fail loudly when stdin is not a TTY, injectable `io.Reader`/`io.Writer`; newline after "Aborted"
- [ ] Commands take their tmux dependency through an interface instead of calling package functions
- [ ] `main.go`: replace `panic` on a config error with stderr + `os.Exit(1)`, handle `Run`'s error, stop shadowing the `cli` package, remove the commented code
- [ ] Prompts, `--yes`, `--dry-run`, `-q` / `-v`
- [ ] Exit codes, stderr/stdout split, `tmxu:` prefix
- [ ] Atomic writes, `0600` files
- [ ] `config`: create the dir on first use, `TMXU_CONFIG_DIR` / `--config-dir`, ignore a relative `XDG_CONFIG_HOME`, docs on `Config`, `DirPath` and `New`
- [ ] Use `cfg.DirPath` in messages and descriptions instead of the literal `~/.config/tmxu`
- [ ] Delete the old top-level `cli/` package once parity is reached

## 8. Decisions

- [ ] One batched chain, separate calls, or `source-file -`
- [ ] Pane names: derived, `@name`, or raw titles
- [ ] On-disk format: reuse these structs or keep a separate persisted shape
- [ ] Existing sessions: skip, refuse, or force
- [ ] Panic or error for a window without panes

## Done in the rewrite

- [x] Typed JSON parsing replaces `SplitN` (no `parts[2]` panics, empty output is `[]`)
- [x] One `list-panes -a` call instead of per-session and per-window calls
- [x] `Session`, `Window`, `Pane` names; `int` indexes; types `Line` and `Data`
- [x] Windows restored at their saved index (no hard-coded base-index for windows)
- [x] Exact-match window and pane targets (`=name:idx`), checked against `man tmux` 3.8
- [x] `%w` wrapping with sentinels per file (except `GetBaseIndex`)
- [x] Package doc comments for `tmux`, `utils`, `cmd` and `config`
- [x] Unused pane sentinels removed; no empty command left in the window chain
- [x] Pure `build*Cmd` functions with unit tests; `Quote` helper with tests; `=` targets covered in `window_test.go`
- [x] `Cmd` interface reduced to `Info()` and `Run(args []string)`, with metadata in one `Info` struct per command
- [x] Package-level `Cmd` vars and the `version` global are gone; commands built by constructors from `config.Config`
- [x] `cmd` types and `config.Config` no longer stutter; command fields unexported and consistent
- [x] `cmd` defines no custom errors (`ListSessions.Run` returns the `tmux` error)
- [x] `tmux.ListSessions` wraps its rows in `[...]` and the `session_windows` format is fixed
- [x] `config.New(version)` trims the version, resolves the dir (XDG on macOS, `os.UserConfigDir` elsewhere) and returns an error
- [x] `version` prints without a trailing space
- [x] Help rendered with `tabwriter`; the per-command boilerplate is gone
