# Review v2: new implementation (`internal/`)

Review of the rewrite in `internal/tmux`, `internal/utils`, `internal/cmd`, `internal/cli` and `internal/config`, compared with the old top-level `cli/` package. Short checklist: [TODOv2.md](TODOv2.md). The original review is [REVIEW.md](REVIEW.md) and its checklist [TODO.md](TODO.md).

Scope: restore functionality first (sections 1 to 7), then the new CLI skeleton (section 8). `main` calls `internal/cli`, but `Run` only prints help; restore is not wired.

Baseline: `go build`, `go vet` and `go test ./...` are clean (86 tests across the repo, last run on commit `0175190`). Claims about tmux behavior that were not run are marked **verify**.

---

## 1. What the new code does

Reads live state with one call, then restores with batched command chains.

```
GetTmuxData()          tmux list-panes -a -F '{json},'   -> []Line -> Data{sessions, windows, panes}
ListSessions()         tmux list-sessions -F '{json},'   -> []Session
CreateSessions()       tmux "new-session -d -s 'a' ; new-session -d -s 'b'"
CreateWindows()        tmux "new-window -k -d -t '=S:1' -n 'w' -c 'dir' ; select-pane -t '=S:1' -T 'title' ; split-window -h ... ; select-pane ..."
GetBaseIndex()         tmux show-options -gv base-index          (not used by anything yet)
```

Pure command builders are separate from the functions that run them (`buildNewSessionsCmd`, `buildNewWindowsCmd`, `buildNewWindowCmd`, `buildNewPanesCmd`, `buildSplitPaneCmd`, `buildRenamePaneCmd`). They are unit tested.

## 2. Old vs new restore flow

| Step                                                      | Old (`cli/`)                                        | New (`internal/tmux`)                              |
| --------------------------------------------------------- | --------------------------------------------------- | -------------------------------------------------- |
| Existing session                                          | `HasSession`; skip, or kill with `-force`           | none                                               |
| Create session                                            | `new-session -d -s NAME -c firstPanePath`           | `new-session -d -s 'NAME'`                         |
| First window                                              | rename the window `new-session` created             | `new-window -k -d -t '=S:idx'` replaces it         |
| Other windows                                             | `new-window -c -t S:idx -n NAME`                    | same command, in one chain                         |
| Extra panes                                               | `split-window -d -c -t S:W`, rename via `S:W.order` | `split-window -h -t '=S:idx'`, `select-pane -T`    |
| Pane name                                                 | derived from window name and order (`paneName`)     | raw `#{pane_title}`                                |
| Layout                                                    | `select-layout` per window                          | none                                               |
| Start dir guard                                           | `startDirArgs` (skips `-c` if not a directory)      | always `-c`                                        |
| Parse tmux output                                         | tab split, index access                             | JSON built in the tmux format string               |
| Orchestration, load/save JSON, templates, attach, prompts | in `cmd.go`, `config.go`                            | none (`ls`, `save`, `version` exist in `internal/cmd`) |

---

## 3. Problems, ranked for restore

### 3.1 `-k` destroys existing windows, and there is no existing-session check

`new-window -k` kills any window already at that index. Restoring onto a live session overwrites its windows and kills their processes. The old code skipped existing sessions. A plain `new-session` on an existing name fails, and tmux stops the chain at the first failing command, so the remaining sessions are never created.

**Fix:** `HasSession` (typed, `(false, nil)` for absent) before creating; skip existing sessions or refuse; add an explicit force path. Never use `-k` against a session that was not created in this run.

### 3.2 Targets: window and pane targets are now exact, the rest is open

Window and pane targets use `'=name:idx'`. `man tmux` (3.8) confirms that an `=` prefix on the session name accepts an exact match only, and that the same applies to windows, so a failed creation can no longer prefix-match another session (`api` vs `api-old`). Still open: the man page says nothing about escaping session names containing `:` or `.`, and `new-session -s` is unaffected.

**Fix:** target by session id (`$3`) once the session exists, or reject names that break target syntax.

### 3.3 Layout is lost

`Line` has no `window_layout`, `Window` has no layout, and every split is `-h`. Multi-pane windows restore as side-by-side columns. A detached `new-session` starts at 80x24 by default, so several `-h` splits can fail with "no space for new pane" (**verify**). The old code applied `select-layout` per window.

**Fix:** read `#{window_layout}`, store it, apply `select-layout` after the splits. Consider `-x`/`-y` on `new-session`, or applying the layout after each split.

### 3.4 First-window replacement assumes saved index == current `base-index`

`new-session` creates one window at the current `base-index`; the chain replaces it only if a saved window has that same index. If the file was saved with `base-index 1` and restored with `0`, the blank window `0` stays and the saved windows are added beside it. `GetBaseIndex` exists but nothing calls it. The old code renamed whichever window `new-session` made.

**Fix:** read the real `base-index` and either map saved indexes onto it, or kill the leftover initial window explicitly.

### 3.5 One bad character fails the whole read

`GetTmuxData` and `ListSessions` build JSON inside the tmux format string without escaping. A `"`, `\` or newline in any session name, window name, path or pane title makes `json.Unmarshal` fail for every row. Programs often put quotes in pane titles, so `save` and `ls` can both fail on one pane.

**Fix:** use a delimiter-based format and split safely, or escape values (tmux has a format modifier for substitution; **verify** the exact syntax on 3.8), or read plain fields per call.

### 3.6 Pane titles are raw `#{pane_title}`

The old code deliberately did not use pane titles (see `paneName` in `cli/tmux.go`) because shells and Claude Code overwrite them. Restored titles will be hostnames or `✳ Claude Code`. I believe `select-pane -T` also expands `#` format sequences, so a title such as `C# #1` could be altered (**verify**). `Quote` does not escape `#`.

**Fix:** decide the naming rule (derived name, or a pane user option such as `@name`), and escape `#` as `##` if tmux expands titles.

### 3.7 Failure mid-chain is opaque

- tmux stops at the first failing command and leaves a half-built session.
- The error is `exit status 1`; stderr is not included and the failing step is unknown.
- No rollback.
- The chain is passed as one `exec.Command` argument. That relies on tmux parsing a single string as a command list (**verify**).

**Fix:** pass `[]string` argv with `";"` elements, or run the script through `source-file -` so tmux reports the failing line. Return a typed error that carries stderr. Roll back sessions created in this run.

### 3.8 Panics instead of errors

`buildNewWindowsCmd` panics when a window has no panes, and `panes[0]` is unchecked. Safe for live data (`parseTmuxLines` always adds a pane with a window), unsafe for windows loaded from JSON. It is the same crash class as REVIEW 1.5, now deliberate.

**Fix:** return an error from the build step, and validate loaded data before building commands.

### 3.9 Order is undefined

`Data.sessions` is a map and `Line` has no `session_created`. The old code sorted sessions by creation time. `list-panes -a` and `list-sessions` return tmux's own order, so `ls` no longer lists oldest first (**verify** tmux's order). `Data.String()` sorts ids as text, so `$10` comes before `$2`.

**Fix:** keep a slice for order, add `session_created`, sort numerically by id or by creation time.

### 3.10 No start-directory guard

The old `startDirArgs` dropped `-c` when the path was not a directory. The new code always passes `-c`. What tmux does with a missing directory was not verified.

**Fix:** port the guard, or verify tmux's fallback and document it.

### 3.11 No persistence design

- `json` tags on `Session`, `Window`, `Pane` carry runtime ids (`$3`, `@7`, `%9`) that mean nothing after a restart.
- `Session` doubles as the tmux DTO: its tags are tmux key names (`session_id`, `session_name`, `session_windows`) while `Window` and `Pane` use `id`, `name`, `index`. A saved file would mix both styles.
- `Window.SessionName` is still stored (REVIEW 4.1 wants derived fields removed).
- No schema version, no migration from the old file, no templates, no `-force`, no current-session guard.

### 3.12 Active pane and window are not restored

After a restore the last split pane is active in each window. `select-pane -T` selects the pane as a side effect (REVIEW 1.7). The original active pane and window are not recorded.

### 3.13 Smaller

- `GetBaseIndex` drops the cause (no `%w`), reads the global option rather than the session's, and is dead code.
- `pane-base-index` is not read. It is not needed now, because panes are addressed through the window's active pane.
- `ErrCannotCreateSessions` has the text "cannot create session".
- `NewSession` hard-codes `WindowsCount: 0` (only `ListSessions` fills it, through JSON).
- `GetTmuxData` and `ListSessions` have no timeout or context, and "no server running" surfaces as `exit status 1`.

---

## 4. Missing features (parity with the old CLI, and new gaps)

Restore:

- Skip existing sessions, `-force`, and refuse to kill the current session.
- Window layout restore.
- Start directory check.
- Stable pane names.
- Initial-window handling that matches the real `base-index`.
- Rollback of partial sessions.
- Order-preserving restore.
- Active pane and window restore.

Persistence and templates:

- Save to and load from the JSON file (`save-sessions` fetches data and discards it), templates, `TEMP_VALUE` handling or its replacement.
- Schema version and migration.
- Atomic writes, `0600` permissions.

CLI surface (`ls`, `save`, `version` exist in `internal/cmd`; the rest only in the old `cli/`):

- `restore-sessions`, `attach`, `new-session`, template commands, menu.
- `help [command]`, dispatch from `Cli.Run`.
- `$TMUX` awareness (`switch-client` inside tmux).
- Exit codes, stderr/stdout split, `-q` `-v`, `--dry-run`.

Infrastructure:

- `run()` helper or `Runner` interface so the exec calls can be faked.
- Typed `TmuxError` with stderr.
- `context` and timeouts.
- Check `tmux -V`, `exec.LookPath("tmux")`.
- "No server running" handled like `tmux ls`.

---

## 5. What the rewrite fixes from the original review

| Original item                                                                                          | Status                                                                                                              |
| ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------- |
| 1.5 Panics from `parts[2]` and `[""]`                                                                  | **Fixed.** Typed JSON replaces `SplitN`; empty output becomes `[]`. `panes[0]` still panics.                        |
| 5.2 Parse output once into typed structs                                                               | **Mostly.** `Line`, one call. No `time.Time`.                                                                       |
| 2.2 Swallowed errors in `tmux.go`                                                                      | **Gone** in the new code, but `HasSession` is not ported.                                                           |
| 1.7 Hard-coded `base-index` / pane index                                                               | **Partial.** Windows use the saved index and pane indexes are unused. First-window handling still assumes the base. |
| 3.2 `t` prefixes, `int16`                                                                              | **Done.** `Session`, `Window`, `Pane`, `Index int`; `cmd.ListSessions`, `config.Config`.                            |
| 3.3 Package docs                                                                                       | **Done** for `tmux`, `utils`, `cmd`, `config`; `internal/cli` is missing. Exported identifiers lost their comments. |
| 2.1 `%w` wrapping                                                                                      | **Done**, except `GetBaseIndex`.                                                                                    |
| 5.3 Table tests for arg building                                                                       | **Partial.** The `build*Cmd` functions are pure and tested.                                                         |
| 4 Drop `SessionName`/`SessionWindow`                                                                   | `Pane` done, `Window` not.                                                                                          |
| 4 `exec.Command` at 12 sites                                                                           | Down to 5. Still no helper or seam.                                                                                 |
| 3.1 Global `Cmd` vars, `version` global, `NewCli` returns an unexported type                           | **Done.** Interface, constructors, exported `Cli`.                                                                  |
| 3.1 Split `cmd.go` per command                                                                         | **Done.**                                                                                                           |
| 7.6 Explicit XDG paths, keep macOS on `~/.config`                                                      | **Done** in `internal/config`. `TMXU_CONFIG_DIR` and `--config-dir` are not.                                        |
| 4 `renderTable` to `tabwriter`                                                                         | **Done** for help; `ls` does not use it.                                                                            |
| 1.4 `HasSession`, 2.3 `TmuxError`, `ErrNoSession`, 1.9 rollback and `-force`, layout items            | **Not addressed.**                                                                                                  |
| `-t =name` exact targets                                                                               | **Partial.** Window and pane targets only.                                                                          |

REVIEW 6.3 recommended keeping restore as separate sequential calls for error attribution. One batched chain gives up that attribution and rollback.

---

## 6. Unidiomatic Go in the new code

- **Stringly-typed commands.** Build `[]string` argv with `";"` elements and call `exec.Command("tmux", args...)`; do not join into one string.
- **Hand-rolled `Comperable` with `Eq` and type assertions.** It allows `Window.Eq(Session)` and compiles. Use `slices.ContainsFunc`, and a `map[string]struct{}` for membership inside loops. `utils.Contain` makes `parseTmuxLines` O(n²) in panes.
- **`Get` prefix and stutter:** `GetTmuxData`, `GetBaseIndex`. Idiomatic: `Snapshot()`, `BaseIndex()`.
- **`GetName`/`GetTitle`/`GetPath`** return a shell-quoted copy of the field, not the field, so the `Get` prefix is misleading (Effective Go: getters drop `Get`). Call `utils.Quote` at the call site, or rename to `Quoted*`. `Window.SessionIndex` is closer to `Target()`.
- **File name** `baseIndex.go` should be `base_index.go`.
- **Dead export:** `GetBaseIndex`.
- **`panic` as error handling** in `buildNewWindowsCmd`, and in `main.go` for a config error.
- **Small:** double map lookup in `parseTmuxLines`, `make(map, 0)`, exported identifiers with no doc comments, `Data` has no exported accessors so other packages cannot use it.

---

## 7. Decisions needed

1. One chain, or separate calls per step? A chain is faster, and separate calls give exact errors. A middle option is `source-file -`.
2. Pane names: derived from the window, `@name` pane option, or raw titles.
3. Do `Data` and its structs become the on-disk format, or only the live snapshot? If on disk, drop ids and `SessionName` and add a version.
4. Existing sessions: skip, refuse, or force with confirmation.
5. Panic or error for a window without panes.

---

## 8. The new CLI skeleton (`internal/cli`, `internal/cmd`, `internal/config`, `main.go`)

### 8.1 What is good

- `cmd.Cmd` is `Info() Info` plus `Run(args []string) error`. Each command keeps its `Info` in a package-level var, compile-time checks (`var _ Cmd = X{}`) guard the interface, and each command lives in its own file.
- `cli.New(cfg)` builds the registry with constructors; aliases are map keys; help uses `tabwriter` over an `io.Writer`.
- `config.New(version)` trims the version, resolves the config dir (`$XDG_CONFIG_HOME` or `~/.config` on macOS, `os.UserConfigDir` elsewhere) and returns an error. `Config{DirPath, Version}` is immutable and passed by value.
- `cmd` defines no custom errors; `ListSessions.Run` returns the `tmux` error.

### 8.2 Problems

1. **`Cli.Run` only prints help.** No dispatch on args, no error return, no exit code, no unknown-command handling, and `help` is not registered as a command.
2. **Help order is random.** `listAllCommands` ranges over the `cmds` map. `cmdsOrder` is filled and never read, and the `seen` set exists only because aliases are also map keys. `help` ignores its `cmdName` argument, hard-codes `os.Stdout` and prints its footer without a newline.
3. **`main.go` panics on a config error** (`panic(fmt.Errorf("%w: cannot read config dir", err))`), repeats the sentinel's text, ignores `Run`'s result and still shadows the `cli` package (`cli := cli.New(cfg)`).
4. **`save-sessions` does nothing useful.** It confirms, calls `tmux.GetTmuxData`, discards the result and writes nothing. Its description and a TODO comment still mention the literal `~/.config/tmxu` and `cfg.Path` (the field is now `DirPath`). `Aborted` is printed without a newline.
5. **`ls` output.** It prints an "Available sessions" header to stdout (pollutes pipes), formats rows with a trailing `\n ` (a stray space on the next line), and does not use `tabwriter`. Order differs from the old command (3.9).
6. **`utils.Confirm`** builds a new `bufio.Reader` per call, ignores the read error, is hard-wired to `os.Stdin`/`os.Stdout`, and silently returns false on EOF (REVIEW 1.9).
7. **Commands call package functions directly** (`tmux.ListSessions`, `tmux.GetTmuxData`), so they cannot be tested without a real tmux server. Inject a small interface.
8. **`Info.Flags` is `[][]string`**, stringly typed. A `Flag{Name, Desc}` struct is clearer.
9. **No alias-collision check** in `newCmd`: a duplicate alias silently overwrites.
10. **Config gaps:** no directory creation, no `TMXU_CONFIG_DIR` or `--config-dir`, a relative `$XDG_CONFIG_HOME` is accepted (the XDG spec requires absolute), no tests (`t.Setenv` makes the env branches easy), and `New`, `Config`, `DirPath` have no doc comments. The darwin branch only runs on macOS.
11. **Doc comments were dropped** from the `cmd` types and constructors in the interface simplification, and `internal/cli` has no package comment.
12. **Old package still present.** The top-level `cli/` (full old implementation) sits next to `internal/cli`; two packages named `cli` are confusing until parity.

### 8.3 Maps to the old TODO

| `TODO.md` item                                              | Status                                                            |
| ----------------------------------------------------------- | ----------------------------------------------------------------- |
| Remove package-level `version` global and `Cmd` vars        | **Done** (the `//go:embed` var in `main` remains)                 |
| Split `cmd.go` per command                                  | **Done**                                                          |
| Decide flat `cli/` vs `internal/`                           | **Decided**: `internal/`                                          |
| Explicit XDG paths (keep macOS on `~/.config`)              | **Done** in `internal/config`                                     |
| Print real resolved paths in messages                       | **Partial**: `cfg.DirPath` exists, messages are still literal     |
| `Run(args) error`                                           | **Signature only**: no dispatch                                   |
| Remove `os.Args` from commands                              | **Started**: commands receive `args`, nothing parses them yet     |
| One `os.Exit` in `main`, exit codes, stderr/stdout          | **Not started**                                                   |
| `confirm` on non-TTY                                        | **Not fixed** (moved to `utils`)                                  |
| `Cmd` custom errors (`ErrCannotListSessionCmd`)             | **Removed**                                                       |
