# TODO v2: `internal/tmux`

Checklist for the rewrite in `internal/tmux` and `internal/utils`. Details and reasoning: [REVIEWv2.md](REVIEWv2.md). The old list for `cli/` is [TODO.md](TODO.md).

Policy: stdlib-first. Write by hand to learn Go. The new code is not wired into the CLI yet.

## 1. Restore safety (fix first)

- [ ] Check `has-session` before creating; skip or refuse existing sessions (today `-k` overwrites live windows)
- [ ] Never use `-k` on a session that was not created in this run
- [ ] Add an explicit force path with confirmation; refuse to kill the current session
- [ ] Exact-match targets: window and pane targets now use `'=name:idx'` (verify tmux accepts it); still open: names containing `:` and `.`, or target by session id
- [ ] Read the real `base-index` (use or delete `GetBaseIndex`); handle saved index != current base (leftover blank window)
- [ ] Roll back sessions created in this run when a later step fails

## 2. Missing restore features (parity with `cli/`)

- [ ] Window layout: read `#{window_layout}`, store it, apply `select-layout` after the splits
- [ ] Decide split direction and the 80x24 detached-size problem (`new-session -x -y`, or apply layout per split)
- [ ] Start directory guard (old `startDirArgs`), or verify tmux's fallback for a missing directory
- [ ] Pane naming rule: derived name, `@name` pane option, or raw title (and escape `#` if tmux expands `-T`)
- [ ] Restore the original active pane and window
- [ ] Preserve session order (`session_created`, slice not map, numeric sort of `$N`)
- [ ] `HasSession` returning `(false, nil)` for absent, typed
- [ ] `attach` / `switch-client` when `$TMUX` is set
- [ ] `list`, `new-session`, templates (`TEMP_VALUE` or its replacement)
- [ ] Load and save the sessions file; templates directory

## 3. Parsing and data model

- [ ] Stop building JSON inside the tmux format (a `"`, `\` or newline in a title or name fails every pane); use a safe delimiter or escaping
- [ ] Add `session_created` (`time.Time`) and `window_layout` to the parsed line
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
- [ ] Finish the quoting-method rename: `Session.StrName` still has the old name, and `GetName`/`GetTitle`/`GetPath` carry a non-idiomatic `Get` prefix; or use one quoting helper at the call site
- [ ] Rename `baseIndex.go` to `base_index.go`
- [ ] Single map lookup in `parseTmuxLines`; drop `make(map, 0)` hints
- [ ] Doc comments on exported identifiers; unexport what is not API

## 6. Testing

- [ ] Update `window_test.go` expectations for the `=` targets (9 tests fail: `TestWindowSessionIndex`, `TestBuildNewWindowCmd`, `TestBuildNewWindowsCmd`)
- [ ] `Runner` interface (`Run`, `Attach`) and `fakeRunner`; stop calling `exec.Command` directly
- [ ] Test `CreateSessions`, `CreateWindows`, `GetTmuxData` through the fake
- [ ] Tests for names with `'`, `"`, `\`, `$`, `#`, spaces, newline
- [ ] Test `parseTmuxLines` ordering and a duplicate-name case
- [ ] Optional integration test on a private socket (`tmux -L tmxu-test-<name>`), skipped without tmux
- [ ] Verify on a throwaway server: single-string chain parsing, `'it'\''s'` quoting, `new-window -k` on a session's only window, `-c` with a missing directory, `-T` format expansion
- [ ] Run tests with `-race`

## 7. CLI surface still in `cli/`

- [ ] Wire `internal/tmux` into `save`, `restore`, `new-session`, `attach`
- [ ] Prompts, `--yes`, `--dry-run`, `-q` / `-v`
- [ ] Exit codes, stderr/stdout split, `tmxu:` prefix
- [ ] Atomic writes, `0600` files, XDG paths
- [ ] Delete the old `cli/tmux.go` once parity is reached

## 8. Decisions

- [ ] One batched chain, separate calls, or `source-file -`
- [ ] Pane names: derived, `@name`, or raw titles
- [ ] On-disk format: reuse these structs or keep a separate persisted shape
- [ ] Existing sessions: skip, refuse, or force
- [ ] Panic or error for a window without panes

## Done in the rewrite

- [x] Typed JSON parsing replaces `SplitN` (no `parts[2]` panics, empty output is `[]`)
- [x] One `list-panes -a` call instead of per-session and per-window calls
- [x] `Session`, `Window`, `Pane` names; `int` indexes
- [x] Windows restored at their saved index (no hard-coded base-index for windows)
- [x] `%w` wrapping with sentinels per file (except `GetBaseIndex`)
- [x] Package doc comments for `tmux` and `utils`
- [x] Types renamed to `Line` and `Data` (no `Tmux` stutter on the types)
- [x] Unused pane sentinels removed from `pane.go`
- [x] No empty command left in the window chain (single-pane windows)
- [x] Pure `build*Cmd` functions with unit tests; `Quote` helper with tests
