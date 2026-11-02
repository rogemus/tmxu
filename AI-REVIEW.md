# AI Review

Reasoning behind [TODO.md](TODO.md). Section numbers match. Covers the move from the old top-level `cli/` package into `internal/` (`tmux`, `utils`, `cmd`, `cli`, `config`) and the work needed to ship `tmxu` as a well-behaved Unix tool.

Policy: stdlib-first, written by hand. `golang.org/x/term` is the only dependency. Single user, so no migration from the old `tmux-sessions.json`. tmux 3.8+ only.

Claims about tmux behavior that were not run are marked **verify**.

---

## 0. Current shape

```
Snapshot (GetTmuxData)  tmux list-panes -a -F '{json},'   -> []Line -> Data{sessions, windows, panes}
ListSessions()          tmux list-sessions -F '{json},'   -> []Session
CreateSessions()        tmux "new-session -d -s 'a' ; new-session -d -s 'b'"
CreateWindows()         tmux "new-window -k -d -t '=S:1' -n 'w' -c 'dir' ; select-pane -t '=S:1' -T 'title' ; split-window -h ... ; select-pane ..."
GetBaseIndex()          tmux show-options -gv base-index   (unused)
```

Pure command builders (`buildNewSessionsCmd`, `buildNewWindowsCmd`, `buildNewWindowCmd`, `buildNewPanesCmd`, `buildSplitPaneCmd`, `buildRenamePaneCmd`) are separate from the functions that run them and are unit tested.

### Old vs new restore flow

| Step                                                 | Old (`cli/`)                                        | New (`internal/tmux`)                                           |
| ---------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------------------- |
| Existing session                                     | `HasSession`; skip, or kill with `-force`           | none                                                            |
| Create session                                       | `new-session -d -s NAME -c firstPanePath`           | `new-session -d -s 'NAME'`                                      |
| First window                                         | rename the window `new-session` created             | `new-window -k -d -t '=S:idx'` replaces it                      |
| Other windows                                        | `new-window -c -t S:idx -n NAME`                    | same command, in one chain                                      |
| Extra panes                                          | `split-window -d -c -t S:W`, rename via `S:W.order` | `split-window -h -t '=S:idx'`, `select-pane -T`                 |
| Pane name                                            | derived from window name and order                  | raw `#{pane_title}`                                             |
| Layout                                               | `select-layout` per window                          | none                                                            |
| Start dir guard                                      | `startDirArgs`                                      | always `-c`                                                     |
| Parse tmux output                                    | tab split, index access                             | JSON built in the tmux format string                            |
| Orchestration, load/save, templates, attach, prompts | `cmd.go`, `config.go`                               | `ls`, `save` (writes nothing), `version`, `attach`, `help` only |

Already fixed by the rewrite (not in the TODO): typed parsing (no `parts[2]` panics, empty output is `[]`), one `list-panes -a` call, `Session`/`Window`/`Pane` names with `int` indexes, no `os.Args` or global `Cmd`/`version`, per-command files, XDG path resolution, `%w` wrapping with per-file sentinels, `Cli.Run(args) error` with one exit site in `main`, `ErrUsage` mapped to exit 2, `-h`/`--help` for every command, `tabwriter` output, injected `IOs`.

---

## 1. Restore safety

### 1.1 `-k` destroys windows; no existing-session check

`new-window -k` kills any window already at that index. Restoring onto a live session overwrites its windows and kills their processes. A plain `new-session` on an existing name fails, and tmux stops the chain at the first failing command, so later sessions are never created. The old code skipped existing sessions.

**Fix:** `HasSession` before creating; skip or refuse existing sessions; an explicit force path; never `-k` a session not created in this run.

### 1.2 `HasSession` treats "no" as an error

`tmux has-session` exits 1 when the session is absent, so the normal "no" case looks like a failure. **Fix:** go through the `run()` helper, `errors.As` to `*exec.ExitError`, treat exit code 1 as `(false, nil)`. Use `if hs` rather than `hs == true`.

### 1.3 Force and the current session

Killing sessions with force can kill the one the user is in, taking their terminal with it. The prompt must list exactly what will be killed. **Fix:** detect the current session (`$TMUX`, `display-message -p '#S'`) and refuse; list targets in the prompt; `-y` to skip; with no TTY and no `-y`, return an explicit error.

### 1.4 Targets

Window and pane targets use `'=name:idx'`; `man tmux` (3.8) confirms the `=` prefix means exact match for the session part, so `api` no longer prefix-matches `api-old`. Still open: the man page says nothing about escaping names containing `:` or `.`, and `new-session -s` is unaffected. **Fix:** target by session id (`$3`) once the session exists, or reject such names. Use `=name` for `has-session`, `attach` and `switch-client` too.

### 1.5 First window and `base-index`

`new-session` creates one window at the current `base-index`. The chain replaces it only if a saved window has that same index. Saved with `base-index 1`, restored with `0`: blank window `0` stays and saved windows are added beside it. `GetBaseIndex` exists but nothing calls it; it also reads the global option, not the session's. **Fix:** read the real `base-index` and map saved indexes onto it, or kill the leftover window explicitly. `pane-base-index` is not needed while panes are addressed through the window's active pane.

### 1.6 Rollback and Ctrl-C

tmux stops at the first failing command and leaves a half-built session that blocks the next run. `CommandContext` kills the child on cancel, so Ctrl-C mid-restore does the same. **Fix:** roll back sessions created in this run; on interrupt, either clean up or report which session was in progress.

---

## 2. Restore parity

### 2.1 Layout

`Line` has no `window_layout`, `Window` has no layout, and every split is `-h`, so multi-pane windows restore as side-by-side columns. A detached `new-session` starts at 80x24, so several `-h` splits can fail with "no space for new pane" (**verify**).

Decisions:

- **tmux 3.8+ only.** `#{window_layout}` is JSON there. Older tmux prints the checksum form (`d2b3,198x53,0,0,1`); not supported, so no dual-form marshaller.
- **Store it as `json.RawMessage`**, a real nested object in the file, not an escaped string. Validate with `json.Valid` on save and fail clearly if tmux returns something else. The layout stays opaque: it is only handed back to `select-layout`.
- **Only for windows with more than one pane** (`json:"layout,omitempty"`; a nil `RawMessage` is omitted). A single pane fills the window, and size comes from the terminal. In the measured file this removes `layout` from 57 of 58 windows.
- **No `select-layout`** for windows without a layout.
- Pane ids inside the layout (`"I":"%0"`) are runtime-only; tmux remaps panes when the layout is applied.
- Apply `select-layout` after the splits; consider `new-session -x -y` or applying per split to avoid the 80x24 problem.

### 2.2 Start directory

The old `startDirArgs` dropped `-c` when the path was not a directory. The new code always passes `-c`. What tmux does with a missing directory is unverified. **Fix:** port the guard, or verify the fallback and document it.

### 2.3 Pane titles

The old code deliberately avoided raw titles (`paneName`) because shells and other programs overwrite them. Restored titles would be hostnames or tool banners. `select-pane -T` may expand `#` sequences (**verify**), and `Quote` does not escape `#`. Also, `select-pane` changes the active pane as a side effect. **Fix:** pick the naming rule (derived name or an `@name` pane option), escape `#` as `##` if needed, and set the title without selecting.

### 2.4 Active pane and window

After restore, the last split pane is active in each window; the original active pane and window are not recorded. **Fix:** record and restore both.

### 2.5 Order

`Data.sessions` is a map and `Line` has no `session_created`. `ls` returns tmux's own order, not oldest-first (**verify**), and `Data.String()` sorts ids as text, so `$10` sorts before `$2`. **Fix:** keep a slice, add `session_created` as `time.Time`, sort numerically.

### 2.6 Inside tmux

`tmux attach -t NAME` fails or nests when `$TMUX` is set. **Fix:** use `switch-client -t =NAME` inside tmux and `attach` outside. Also lets `new-session` behave like `tmux new -A` if wanted.

---

## 3. Parsing and data model

### 3.1 One bad character fails everything

`Snapshot` and `ListSessions` build JSON inside the tmux format string with no escaping. A `"`, `\` or newline in any name, path or pane title makes `json.Unmarshal` fail for every row; programs often put quotes in titles. **Fix:** a delimiter-based format split safely, or escaping (tmux has a substitution modifier; **verify** syntax on 3.8), or plain fields per call.

### 3.2 Persisted shape

Measured on a real file from the old tool: 15 sessions, 58 windows, 59 panes, 33,291 bytes.

| Field                           | Redundant in                     | Derivable from              |
| ------------------------------- | -------------------------------- | --------------------------- |
| `window.sessionName`            | 58/58                            | `session.name`              |
| `window.sessionWindow`          | 58/58                            | `session.name:window.order` |
| `pane.sessionName`              | 59/59                            | `session.name`              |
| `pane.sessionWindow`            | 59/59                            | parent window               |
| `order`                         | 132/132 equal to array index + 1 | array index                 |
| `pane.name`                     | 57/59 equal to `window.name`     | `window.name`               |
| `layout` on single-pane windows | 57/58                            | not needed                  |

Dropping these shrinks the file about 46% (33,291 to 17,841 bytes) without changing restore behavior. Duplicated facts are only valid while every copy agrees, and are why templates needed the `TEMP_VALUE` hack. With no legacy data to keep, the new file is designed clean:

```json
{
  "version": 1,
  "sessions": [
    {
      "name": "myproject",
      "windows": [
        { "name": "editor", "panes": [{ "path": "/Users/me/Developer/project" }] },
        {
          "name": "dev",
          "layout": { "V": 2, "L": { "...": "..." } },
          "panes": [{ "path": "/a" }, { "path": "/b" }]
        }
      ]
    }
  ]
}
```

- Keep `pane.path` explicit per pane; inferring it from the window is fragile.
- Keep `order` only if real tmux indexes (gaps from closed windows) must survive; otherwise use the array index plus the real `base-index`.
- Keep the `"version"` field. It costs one line and makes a future format change detectable. No migration code.
- Drop runtime ids (`$3`, `@7`, `%9`) and `Window.SessionName`.
- A template is a `Session` with no runtime fields, so `TEMP_VALUE` goes away and `save` and `save-template` share one `snapshotSession`.

### 3.3 DTO vs domain

`Session` doubles as the tmux DTO: its tags are tmux key names (`session_id`, `session_name`, `session_windows`) while `Window` and `Pane` use `id`, `name`, `index`. A saved file would mix both styles. **Fix:** a parsed-line type for tmux output, a domain type, and a persisted type; recommended separate persisted shape.

### 3.4 Panics

`buildNewWindowsCmd` panics for a window without panes and `panes[0]` is unchecked. Safe for live data, unsafe for loaded JSON. **Fix:** return an error from the build step and validate loaded data first. `NewSession` also hard-codes `WindowsCount: 0`.

---

## 4. Errors and execution

### 4.1 Execution seam

`exec.Command` is called directly in 5 places; nothing is fakeable. Introduce one seam, and put `-v` logging and `--dry-run` behind it:

```go
type Runner interface {
    Run(ctx context.Context, args ...string) ([]byte, error) // capture stdout
    Attach(ctx context.Context, args ...string) error        // inherits stdio
}
```

Commands receive a small interface (defined where it is consumed) instead of calling package functions, so they test without a tmux server.

### 4.2 One string vs argv

Joining the chain into a single `exec.Command` argument relies on tmux parsing one string as a command list (**verify**). **Fix:** pass `[]string` argv with `";"` elements, or run a script via `source-file -` so tmux reports the failing line. Decision pending: one chain (fast, opaque errors), separate calls (exact errors, more forks), or `source-file -`.

### 4.3 Typed errors

```go
type TmuxError struct {
    Args   []string
    Stderr string
    Err    error
}
func (e *TmuxError) Error() string { ... }
func (e *TmuxError) Unwrap() error { return e.Err }
```

Today failure is `exit status 1` with no stderr and no failing step. Add `ErrNoSession`, `ErrNoTemplate`, `ErrNoTmux` (`errors.Is(err, exec.ErrNotFound)` → "tmux not found in PATH"). `GetBaseIndex` drops its cause (no `%w`), `ErrCannotCreateSessions` says "cannot create session". Use `errors.Is(err, fs.ErrNotExist)` and call `os.Remove` directly instead of `os.Stat` + `os.IsNotExist` (check-then-act, TOCTOU). Use `errors.Join` for multiple restore errors.

### 4.4 No server, no tmux

`ls` with no server should behave like `tmux ls`, not fail with `exit status 1`. Check `exec.LookPath("tmux")` and `tmux -V` at startup with clear messages.

### 4.5 Context and signals

No `context` anywhere. `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` in `main`; `exec.CommandContext` with a 5 s timeout for non-interactive calls so a hung server socket cannot block the CLI; no timeout for `attach` (interactive, long-lived). Exit 130 on interrupt.

### 4.6 Concurrency

No goroutines today, so no races.

- **Safe (read-only):** snapshotting windows and panes per session. Write results by index (`out[i]`, `errs[i]`), bounded by a channel semaphore, so no mutex and deterministic order. Each tmux call is about 5-10 ms of fork, so it only matters with dozens of windows. Measure first; good learning exercise. Run under `-race`.
- **Risky:** restoring independent sessions in parallel. The first `new-session -d` starts the server and racing starts can fail; create the first sequentially. Interleaved output, error policy.
- **Must stay sequential:** window and pane creation inside one session (indexes exist only after the previous call), `NewSession` then first-window handling then panes, pane renames, the menu and `attach`.

---

## 5. Port the remaining commands

### 5.1 `save` does nothing

It confirms, fetches data, discards it and writes nothing. Its description still mentions the literal `~/.config/tmxu`. **Fix:** write the file through the config store; share `snapshotSession` with `save-template`; share `restoreSession` between `restore-sessions` and `new-session -templ`; one helper maps `errSessionExists` to a user message.

### 5.2 Input validation

- `new-session` with an empty name must be rejected.
- Template names are used in paths, so `save-template -name ../x` would write outside the templates directory. **Fix:** `validName` rejects empty, `/`, `\`, `..`, leading `.`.
- Loading templates must skip non-`.json` entries and subdirectories, and report a bad file without failing the whole list.
- `delete-template`: clear error for a missing template, `-f` for `rm` semantics.

### 5.3 Menu (`attach -menu`)

Not ported; the flag is parsed but never read and its example has a typo. If implemented, the old menu's defects must not come along:

- `term.MakeRaw` and read errors were ignored (a non-TTY passes a nil state to `Restore`; EOF is a 100% CPU loop).
- In raw mode Ctrl-C is byte `0x03`, not SIGINT, and nothing handled it.
- `strings.Repeat("·", 25-len(title))` panics for long titles and counts bytes, not runes.
- Enter on empty items indexed out of range.
- It cleared the whole screen (destroys scrollback) with repeated escape literals.
- Esc detection through `Buffered() > 0` is racy.

**Fix:** `term.IsTerminal` on stdin and stdout with a plain-list fallback; handle EOF and `case 3, 4`; constants for escapes and the alternate screen (`\033[?1049h`/`l`); draw on stderr or `/dev/tty` so stdout stays clean for the selection; multibyte-safe padding. Defers do not run on a fatal signal, so return through normal control flow on context cancellation; the blocking read needs a goroutine or an fd close to be interruptible.

### 5.4 `utils.Confirm`

New `bufio.Reader` per call, read error dropped, EOF silently means "no", so cron or piped use prints "Aborted" and exits 0. **Fix:** one reader, handle the error, fail with "refusing to prompt: stdin is not a terminal; pass --yes". Rename the `err io.Writer` parameter to `out`.

### 5.5 Attach plumbing

`AttachToSession` writes to `os.Stdin/Stdout/Stderr` instead of the injected `IOs`, and its error format `%w %s: %w` lacks a colon.

### 5.6 Config

- Files hold working-directory paths and window names: `0600` files, `0700` dirs.
- `os.WriteFile` truncates in place; a crash mid-write corrupts the only backup. One helper: `os.CreateTemp` in the same dir, write, `Sync`, `Close`, `Chmod`, `Rename`. Keep a `.bak`.
- Create the directory on first use (`os.MkdirAll`).
- `TMXU_CONFIG_DIR` and `--config-dir`, precedence flag > env > default. Do not clear the environment (`TMUX_TMPDIR` and `TMUX` are consumed by tmux).
- Ignore a relative `XDG_CONFIG_HOME` (the spec requires absolute).
- Keep macOS on `~/.config`: `os.UserConfigDir` would return `~/Library/Application Support`.
- Print `cfg.DirPath` in messages, not the literal `~/.config/tmxu`.
- `readJSON[T any]`, `writeJSON`, `filepath.Join` everywhere.
- Doc comments on `Config`, `DirPath`, `New`.

### 5.7 Cut-over

Two packages named `cli` are confusing. Delete the top-level `cli/` as soon as parity is reached.

---

## 6. Unix-style CLI surface

### 6.1 Flags and help

stdlib `flag` has no short/long pairs and stops at the first positional, so `tmxu new mysession -path /tmp` silently ignores `-path`. **Fix:** document "flags before arguments" and detect stray `-x` tokens, or pre-split args into flags and positionals yourself; honor `--` so a session named `-x` works. Add `-V`/`--version`. Help layout: Usage, Commands, Options, Environment, Exit status; print `Examples`. `Info.Flags` as `[][]string` is stringly typed (`index . 0` in the template); use `Flag{Name, Desc}` and derive from `FlagSet.VisitAll`. Detect alias collisions in `newCmd` (a duplicate alias silently overwrites). Unknown command: suggest the nearest (small Levenshtein, about 15 lines). Missing required argument: usage and exit 2. Decide what bare `tmxu` does.

### 6.2 Exit codes and streams

0 success, 1 runtime failure, 2 usage error, 130 interrupt (consider 127 for "tmux not installed"). Only data on stdout; errors, prompts and "Aborted" on stderr; every error prefixed `tmxu:`. Declining a prompt in a non-interactive run should exit 1 so scripts can tell.

### 6.3 Quiet, verbose, dry-run, yes

`-q` suppresses "Session created..."; `-v` logs each tmux command (trivial with the runner) via `log/slog` on stderr; `-n/--dry-run` prints actions instead of doing them; `-y/--yes` skips prompts. Restore without force is idempotent per session (skips existing); document it.

### 6.4 Machine-readable output and pipelines

`ls` and `list-templates`: plain one-name-per-line when stdout is not a TTY, `--json`, `-0` for `xargs -0`. `attach -` reads a name from stdin so `tmxu ls | fzf | tmxu attach -` works. `save -o -`, `restore -f FILE`, `show-template NAME`. `tmxu paths` prints resolved directories for scripts and bug reports.

### 6.5 Terminal behavior

TTY detection via `x/term`; `NO_COLOR` and `TERM=dumb` respected if color is added; handle SIGPIPE/EPIPE so `tmxu ls | head -1` is quiet; multibyte-safe padding. tmux is Unix-only: `//go:build !windows` or a clear unsupported-OS message.

### 6.6 Idiomatic Go cleanup

- `Comperable`/`Eq` with type assertions allows `Window.Eq(Session)` and compiles. Use `slices.ContainsFunc` and a `map[string]struct{}`; `utils.Contain` makes `parseTmuxLines` O(n²) in panes.
- `GetTmuxData`/`GetBaseIndex`: `Get` prefix and stutter → `Snapshot()`, `BaseIndex()`.
- `GetName`/`GetTitle`/`GetPath` return a shell-quoted copy, not the field. Call `utils.Quote` at the call site, or name them `Quoted*`.
- `baseIndex.go` → `base_index.go`.
- `wantsHelp` is a method that ignores its receiver; the `cfg` field is unused in three commands.
- Exported identifiers lost their doc comments; `internal/cli` has no package comment; unexport what is not API; `Data` has no exported accessors.
- Single map lookup in `parseTmuxLines`.

---

## 7. Testing

Current tests cover the pure `build*Cmd` functions and `Quote`. Nothing runs through a fake, nothing covers `internal/cli` or `config`, and `-race` is not used.

- **Fake runner** with `calls [][]string` and `fn func(args []string) ([]byte, error)`; assert exact args.
- **Tmux layer:** `CreateSessions`, `CreateWindows`, `Snapshot`, `ListSessions` ordering, empty output, a window without panes.
- **Hostile names:** `'`, `"`, `\`, `$`, `#`, spaces, newline.
- **Config:** a store with an injected directory (`t.TempDir()`); round trip, missing file, malformed JSON, traversal name (must be rejected), non-`.json` file, delete missing. `config.New` with `t.Setenv`.
- **`internal/cli`:** help output order and aliases, command lookup, `ErrUnknownCommand`, `-h` handling.
- **Command level:** injected stdin/stdout/stderr plus the fake runner run `save`, `restore`, `new -templ` end to end.
- **Golden files** in `testdata/` with an `-update` flag: saved JSON (pins the schema), help output, `ls` table.
- **Integration (optional):** private socket `tmux -L tmxu-test-<name>`, skipped when tmux is missing. The runner needs to accept global prefix args.
- **Verify on a throwaway server** (these decide the design): single-string chain parsing, `'it'\''s'` quoting, `new-window -k` on a session's only window, `-c` with a missing directory, `-T` format expansion, order of `list-panes -a`.
- Helpers call `t.Helper()`; temp state uses `t.Cleanup`; run `go test -race ./...`.

---

## 8. Ship it

### 8.1 CI and release

- No build, vet, test or lint runs on PRs today. `ci.yaml`: `go build ./...`, `go vet ./...`, gofmt diff check, `go test -race ./...`, staticcheck, `go mod tidy` diff. Pin actions by SHA (`actions/checkout@v5` is unpinned); Dependabot for modules and actions.
- The existing workflow tags `version.txt` on push to main. `git tag` fails when the tag exists, there is no `git config user.name/email`, no `v*` check, and no GitHub Release, binaries, checksums or changelog. The version lives in two places.
- **Fix:** trigger on `push: tags: ['v*']`; the git tag is the single source: `-ldflags "-s -w -X main.version=$TAG"` with `-trimpath` and `CGO_ENABLED=0`, falling back to `debug.ReadBuildInfo().Main.Version` for `go install` users. Delete `version.txt`. Matrix darwin/linux × amd64/arm64, `sha256sum`, `gh release create`.
- Alternative if you later value speed over learning: goreleaser (archives with LICENSE/README/man/completions, checksums, changelog, Homebrew tap).

### 8.2 Completion

`tmxu completion zsh|bash|fish` prints a script to stdout; embed with `//go:embed` or generate with `text/template` from the command registry so it cannot drift. Dynamic names need a hidden `tmxu __complete <cmd> <prefix>` that prints session names (`tmux list-sessions -F '#{session_name}'`) for `attach`/`save-template`, and template names (strip `.json`) for `delete-template` and `new -templ`. Test it with the fake runner and a temp config dir. Install through the formula or release script.

### 8.3 Man page

`man/tmxu.1` in roff: NAME, SYNOPSIS, DESCRIPTION, COMMANDS, OPTIONS, FILES, ENVIRONMENT, EXIT STATUS, EXAMPLES, SEE ALSO (tmux(1)). Install with `man1.install` in the formula. `go install` users get no man page, so keep `--help` complete and say so in the README. Write it after the flags stop changing.

### 8.4 Docs and repo hygiene

- **README:** Go "1.16+" is wrong (`go.mod` requires 1.26). Add the tmux 3.8+ requirement, install, completion setup, exit codes, JSON format, an fzf recipe, and how it differs from tmux-resurrect, tmuxinator and sesh.
- **`.gitignore`:** the unanchored `tmxu` also ignores any future `cmd/tmxu/`; use `/tmxu`. Add `.DS_Store`, `dist/`, `.claude/`, `*.test`, `coverage.out`.
- **`justfile`:** `run` passes the literal string rather than the variable (needs `{{ARGS}}` with `*ARGS`); mixed tabs and spaces; `lint` runs `go fix -diff`, which is not a linter. Add `build`, `install`, `check` (fmt + vet + staticcheck + `test -race`), groups.
- **`CLAUDE.md`** is stale (`gh.go`, `Formula/tmxu.rb`, old type names, `cli/` layout): fix it or leave it untracked.
- **Add** `CONTRIBUTING.md` (build with `just`, run tests, commit convention, release by tagging), `CHANGELOG.md`, `.editorconfig`.
- **Homebrew:** no `Formula/` directory exists, so references are dead. Add a tap that installs the man page and completions, or remove the references.

---

## 9. Optional: restore performance (parked)

Symptom: after `tmxu restore`, attaching shows panes catching up (prompts redraw, shells still starting).

Measurement: `time zsh -i -c exit` is 0.127 s (about 60 ms CPU), so one shell is fast. The cost is aggregate: 59 panes × ~60 ms ≈ 3.5 s of CPU contending across cores, and the user attaches while many are still booting. Not yet measured: `time tmxu restore`; take that baseline first. If it is under about 2 s, items 1 and 2 below should be enough.

Likely causes (from reading the old code; not profiled):

1. **Detached size, then resize on attach.** `new-session -d` is 80x24, `select-layout` applies at that size, and attaching resizes every pane (SIGWINCH redraw).
2. **Shells still booting.** Restore returns before prompts are drawn.
3. **Many forks.** Several hundred `exec.Command` calls; real but small next to the shells.

Ideas by value:

1. **Wait for readiness (main fix).** After creating all panes, poll `capture-pane -p -t <pane>` every ~50 ms until non-empty, about 5 s timeout per session, sessions in parallel (`sync.WaitGroup`) so the wait is the slowest shell, not the sum. `waiting for shells...` on stderr; on timeout warn and continue; `-no-wait` skips it. Read-only polling, a good goroutine exercise. A shell with an empty prompt never produces output and hits the timeout.
2. **Real size.** `new-session -x <cols> -y <rows>` from `term.GetSize(int(os.Stdout.Fd()))`, no flags when not a TTY. Only helps when the attaching terminal has the same size.
3. **Batch calls** per window into one invocation (low priority).
4. **Stagger** in batches of 3-4 sessions (only if still janky).

Alternative readiness signal: `send-keys 'tmux wait-for -S <name>' Enter` is exact but echoes in the pane; prefer `capture-pane`. Decision: wait by default with `-no-wait` (recommended) or opt-in `-wait`.

---

## 10. Decisions needed

1. One chain, separate calls, or `source-file -`: chain is fastest, separate calls give exact errors, `source-file -` is the middle.
2. Pane names: derived from the window, `@name` option, or raw titles.
3. On-disk format: reuse the `internal/tmux` structs or a separate persisted shape (recommended: separate).
4. Existing sessions on restore: skip, refuse, or force with confirmation.
5. Panic or error for a window without panes (recommended: error).
6. Wait for shells by default or opt-in.

---

## Suggested order

1. **Seam and parsing:** `Runner`, `TmuxError`, safe line format, typed `Line`. Concepts: small interfaces, fakes.
2. **Restore safety and parity** with tests on the fake: `HasSession`, no `-k`, `base-index`, layout, rollback.
3. **Persisted shape and config store:** clean schema, atomic writes, `0600`. Concepts: `os.CreateTemp`, `os.Rename`, generics for `readJSON[T]`.
4. **Port commands**, then delete `cli/`.
5. **`context` and signals.** Concepts: `exec.CommandContext`, `signal.NotifyContext`.
6. **CLI surface:** flags, exit codes, `--json`, `-0`, `-y`, `--dry-run`.
7. **Goroutines:** parallel snapshot and shell-wait polling, under `-race`.
8. **Ship it:** CI, release, completion, man page, docs.
