# TODO v2: new implementation (`internal/`)

Checklist for the rewrite in `internal/tmux`, `internal/utils`, `internal/cmd`, `internal/cli` and `internal/config`. Details and reasoning: [REVIEWv2.md](REVIEWv2.md). The old list for the top-level `cli/` package is [TODO.md](TODO.md).

Policy: stdlib-first. Write by hand to learn Go. `main` is wired to `internal/cli`; `Run` dispatches `ls`, `save`, `version` and `help`; restore is not wired.

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

- [ ] Wire `internal/tmux` into `restore`, `new-session`, `attach` (`ls` and `save` call it; `save` writes nothing)
- [ ] `Info.Flags` as `[]Flag{Name, Desc}` instead of `[][]string`
- [ ] Detect alias collisions in `newCmd`
- [ ] `utils.Confirm`: one reader, handle the read error, fail loudly when stdin is not a TTY
- [ ] `utils.Confirm`: rename the `err io.Writer` param to `out`
- [ ] Commands take their tmux dependency through an interface instead of calling package functions
- [ ] `attach`: inside tmux (`$TMUX` set) use `switch-client` instead of `attach`; target with `=name` (exact match)
- [ ] `attach`: implement `-menu` (fzf-style menu) or remove the flag and its example; fix the `tmux attach-session -menu` typo in `Examples`; drop the blank first line of `Run`
- [ ] `tmux.AttachToSession` writes to `os.Stdin`/`Stdout`/`Stderr`; take the `cmd.IOs` streams instead; error format `%w %s: %w` is missing a colon
- [ ] `wantsHelp` is a method that does not use `c`; make it a plain function
- [ ] Unused `cfg` field in `AttachSession`, `ListSessions`, `Version`
- [ ] Prompts, `--yes`, `--dry-run`, `-q` / `-v`
- [ ] Tests for `internal/cli` (help output, command lookup, `ErrUnknownCommand`)
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
- [x] `cmd` defines one shared `ErrUsage` plus the `attach` sentinels; `ListSessions.Run` returns the `tmux` error
- [x] `tmux.ListSessions` wraps its rows in `[...]` and the `session_windows` format is fixed
- [x] `config.New(version)` trims the version, resolves the dir (XDG on macOS, `os.UserConfigDir` elsewhere) and returns an error
- [x] `version` prints without a trailing space
- [x] Help: command list with `tabwriter` in `cmdsOrder`; per-command help (aliases, description, usage, flags, examples) from a `text/template`; both write to an injected `io.Writer`
- [x] Commands get `cmd.IOs{In, Out, Err}` by constructor; `version` and `ls` write to `Out`, prompts and "Aborted" to `Err`
- [x] `Cli.Run(args) error`: dispatch, `help [command]`, `ErrUnknownCommand` through one `unknownCmdError` helper; consistent `*Cli` receivers, no shadowing of the `cmd` package
- [x] `ls`: no header, `tabwriter` with `Flush`
- [x] `main.go`: one exit site, exit 1 for runtime errors and 2 for any `cmd.ErrUsage` (unknown command, bad flag, missing session name), `tmxu:` prefix on stderr, no `panic`, no `io` shadowing
- [x] `-h`, `-help`, `--help` handled in `Cli.Run` for every command (help to stdout, exit 0); `tmxu -h` and `tmxu --help` print general help; scanning stops at `--`
- [x] `attach-session`: flags parsed with a silent `FlagSet`, session name checked after `Parse`, `tmux attach -t` wired to the terminal
- [x] `utils.Confirm` takes an `io.Reader` and writes the prompt to stderr; "Aborted" goes to stderr with a newline
