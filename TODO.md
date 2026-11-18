# TODO

Everything left to do to finish porting `restore` into `internal/` and to ship `tmxu` as a good Unix tool.

Policy: stdlib-first. Only `golang.org/x/term` is allowed as a dependency (currently not in `go.mod`; add it with the menu or `term.GetSize` work). Write things by hand to learn Go.
Scope: single user, so there is no migration from the old `tmux-sessions.json`. The new schema is designed clean.
Target: tmux 3.8+ only (JSON `window_layout`).

Status: `main` is wired to `internal/cli`. `Run` dispatches `ls`, `save`, `version`, `help`, `attach`, `dump`, `new-session`, `restore` (registered, `Run` is still a stub), `list-templates`, `save-template`, `delete-template`, `show-template`, `paths`. Config file operations go through `os.Root` (helpers take a `*os.Root` and a root-relative name, and return `filepath.Join(root.Name(), name)`); this covers file reads, writes and deletes only, directory creation and `ListTemplates` still use plain paths. Each command owns its `FlagSet` (built in the constructor, bound to its own fields, pointer receivers); `Cmd.Flags()` is the single source of flags for parsing and help. Restore and the menu are not ported. `new-session -templ` works end to end (error paths are still open, see Bugs). `dump` and `show-template` print a names-only tree (session, windows, panes with their paths); `-json` works on `ls`, `list-templates`, `version`, `paths`, `dump` and `show-template`; `save` has `-yes`. `save` writes `tmux-sessions2.json` into the config dir (first version: not atomic, files 0600, dirs 0700). Window layouts are read, stored as raw tmux JSON and applied after the splits. `internal/` tests: only `utils` (`Quote`, `Escape`, `Slug`) plus the `internal/expect` helper, all written by hand (see Tests). Everything that shells out to tmux needs the runner seam.

Done items are removed from this file; `git log` has them. Goal: 0.8.0 release. Port `restore` to `internal/` (the old `cli/` package is already deleted). `attach -menu` is not ported yet: it was broken in the previous version. The flag stays registered (and `restore -force`) and will be implemented.

Sections: [Bugs](#bugs) · [Cosmetic changes](#cosmetic-changes) · [Features](#features) · [Tests](#tests) · [Nice to have](#nice-to-have) · [Decisions](#decisions)

## Bugs

Wrong behavior, reachable panics, silent failures, and unsafe states in code that exists today.

### Templates and `new-session -templ`

1. No rollback: a failed window step leaves a half-built session that blocks the retry with `duplicate session` (also applies to restore, see Features)
2. `--path` is not applied to template windows or panes (`SanitizeSnapshot` clears every `Pane.Path` and the `-c` flag is skipped for an empty path); see Decisions
3. Session names with `.` or `:` (tmux may rewrite them, so the printed name and `attach` could disagree; verify, then reject, print the real name, or target by session id)

### Layout and tmux commands

1. `CreateSessions` and `buildNewSessionsCmd` (`session.go`) have no callers and still pass one joined string to `exec.Command`; delete them, or rewrite on argv when restore needs them
2. Stop building JSON inside the tmux format string (`Snapshot`/`SessionSnapshot` through `buildListPanesFormat`, `ListSessions`): a `"`, `\` or newline in a name or title fails every row; use a safe delimiter or escaping

### Sessions

1. `HasSession` (`session.go:102-113`) returns an error on every non-zero exit, decides by empty stdout and uses `-t NAME` (not `=NAME`); it should return `(false, nil)` for an absent session (exit code 1 from `has-session`) and a typed error for real failures
   - Decision: exit 1 means `(false, nil)`, which also covers "no server running"; other exit codes and `exec.ErrNotFound` stay typed errors
   - Verified: absent session exits 1 with `can't find session: NAME`; no server exits 1 with `error connecting to ...`; success exits 0 with empty stdout, so decide by exit code (`errors.As` + `*exec.ExitError`), not by output
   - Use `-t "=NAME"` (exact match; `list-panes -s -t =NAME` verified); quote `=NAME` in a zsh shell, it expands otherwise
   - `save-template` must return an explicit "session not found" error for `false` (`save_template.go:72-75`: `err != nil || !hasSession { return err }` returns nil)
   - `dump -session` targets `=NAME` (`tmux.go:97`) but `HasSession` does not; make them agree

### CLI behavior

1. Flags after positional args work; `--` terminator honored (today `-path` after a name is silently ignored)
2. Only `new-session` rejects extra positional args (`ErrUnexpectedArgument`); `dump`, `ls`, `version`, `paths`, `save`, `attach`, `save-template`, `show-template` and `delete-template` ignore them
3. `utils.Confirm`: one reader, handle the read error, fail loudly when stdin is not a TTY

## Cosmetic changes

Naming, text, structure, duplication, doc comments. No behavior change for users.

### Naming and text

1. `helpCmd`: align the FLAGS descriptions in a column (`-path` and `-templ` differ in length, the template pads with fixed spaces); use `text/tabwriter` like `listAllCommands`, with a tab between name and description, flushed after `Execute`; check that the other sections are not affected
2. Doc comments on exported identifiers (package docs exist everywhere; `Info`, `Flag`, `IOs`, `Cmd`, `Config`, `New`, `Session`, `Window`, constructors and `utils.Slug` have none); unexport what is not API

### Duplication and dead code

1. `utils.Quote` has no callers on the argv path; only `buildNewSessionsCmd` (dead, `GetName` is gone) uses it (`printError` is gone); delete it with `CreateSessions` and `buildNewSessionsCmd`. Note `new-session -s` is format-expanded, so a rewrite on argv needs `utils.Escape` for the name
2. `save`: one shared `snapshotSession` for sessions and templates
3. `Snapshot` and `SessionSnapshot` duplicate about 10 lines (see the shared `snapshotSession` item)
4. `HasTemplate` duplicates what `ShowTemplate` reports; keep it, or return a not-found sentinel from `ShowTemplate`
5. Config: fold the repeated `os.OpenRoot` + `defer root.Close()` (5 places: four in `templates.go`, one in `sessions.go`) into one helper
6. Generic `readJSON[T]` / `writeJSON`; `os.MkdirAll` (file reads, writes and deletes already go through `os.Root`: `readFile`, `saveToFile`, `deleteFile` in `tmxu_config.go`)
7. `Root`-based listing for `ListTemplates` (`fs.ReadDir(root.FS(), ".")`; `os.Root` has no `ReadDir`; `readTemplatesName` then takes `[]fs.DirEntry`)

### Structure and refactors

1. Rebuild the `json:"-"` fields after loading the file (`Window.Slug`, `Window.TargetIndex` from the session name, window name and index), in one `Data` method so the formula lives once (the formulas now live in `sessionSlug`, `windowSlug` and `windowTarget`, but `SanitizeSnapshot` and `DesanitizeTemplSnapshot` still rebuild the keys by hand; a loaded snapshot has an empty `Window.Slug`, which `SnapshotData.String` works around with `windowSlug`, so any other reader of `w.Slug` after `LoadSnapshot` still finds nothing)
2. Templates are a `Session` without runtime fields; the `TEMP_VALUE` placeholder is now the empty session name (`templateSessionName`), but `SanitizeSnapshot` and `DesanitizeTemplSnapshot` still re-key every map; remove them, and the constant, with the persisted-shape work
3. Persisted shape (clean schema, `"version": 1`, no migration): `{"version":1,"sessions":[{"name","windows":[{"name","layout"?,"panes":[{"path"}]}]}]}`; drop the redundant and runtime fields (`sessionName`, `sessionWindow`, runtime ids, `order` = array index, `pane.name` = window name, `layout` on single-pane windows; about 46% smaller); a separate persisted type instead of reusing the tmux structs; a template is a `Session` with no runtime fields, so `save` and `save-template` share `snapshotSession`
4. One `run()` helper or a `Runner` interface (`Run`, `Attach`) with an `execRunner`; no direct `exec.Command` elsewhere
5. Commands take their tmux dependency through an interface instead of calling package functions
6. Typed `*TmuxError{Args, Stderr, Err}`; include stderr in messages
7. Sentinels `ErrNoSession`, `ErrNoTmux` (`exec.ErrNotFound`); "tmux not found in PATH" message (`ErrNoTemplate` lives in `cmd`, `ErrSessionExists` in `tmux`)
8. `errors.Join` to collect restore errors

## Features

New capabilities and the remaining migration. Restore comes first, then the CLI surface, then release work.

### Migration

1. `restore-sessions` (`restore`, `r`), one shared `restoreSession` for restore and new-from-template
   - `-force`

### Restore safety

1. Check `has-session` before creating; skip or refuse existing sessions (today only `NewSession` detects a duplicate, from tmux's stderr; a duplicate in a `CreateSessions` chain stops the rest)
2. Never use `-k` on a session that was not created in this run
3. Explicit force path with a prompt that lists what will be killed; refuse to kill the current session (`$TMUX`, `display-message -p '#S'`)
4. Read the real `base-index` (`BaseIndex` was deleted as unused; `tmux show-options -gv base-index` reads the global option, the session's own may differ); handle saved index != current base so no blank leftover window stays
5. Roll back sessions created in this run when a later step fails
6. Ctrl-C during restore: clean up or report the partial session

### Restore parity with the old `cli/`

1. Decide split direction and the 80x24 detached-size problem (`new-session -x -y`, or apply layout per split)
2. Start directory guard (old `startDirArgs`): verified that tmux does not fail on a missing `-c` directory, it silently starts in `$HOME`; decide whether to warn
3. Restore the original active pane and window
4. `attach` / `switch-client` when `$TMUX` is set; `=name` exact target

### Errors, signals, environment

1. "No server running" handled like `tmux ls`
2. `exec.LookPath("tmux")` and `tmux -V` check at startup with a clear message
3. Thread `ctx`; `signal.NotifyContext` in `main`; `exec.CommandContext` with a 5s timeout for non-interactive calls, none for `attach`
4. Exit code 130 on interrupt (0, 1, 2 done)

### Templates and config

1. Templates: clear missing-file error for `show-template` / `delete-template`, `-f` for `delete-template`, and an overwrite policy for `save-template` (it truncates silently, see Decisions)
2. Config store: `writeFileAtomic` (temp file, `Sync`, `Rename`, same `0600` mode), keep a `.bak`
3. Config: `TMXU_CONFIG_DIR` env and `--config-dir`; ignore a relative `XDG_CONFIG_HOME`; doc comments on `Config`, `BaseDirPath`, `AppDirPath`, `New`

### CLI surface

1. Help layout: Usage, Commands, Options, Environment, Exit status
2. `-yes` for restore, new and delete-template (done for `save-sessions`), `-n` / `--dry-run` (restore, new, delete-template, save), `-q` / `-v` with `log/slog` to stderr (`-v` logs tmux args)
3. Decide what bare `tmxu` does (help + exit 0 today)

### Ship it (after the migration)

1. Switch `save` back from `tmux-sessions2.json` to `tmux-sessions.json` (temporary name while the port is in progress); change only `sessionsFileName` (`paths` and the `DescLong` text follow it) and update the status text
2. CI: `main-check.yaml` already runs build, vet, gofmt and `go test -race ./...`; add staticcheck and a `go mod tidy` diff
3. Release on `v*` tag push; git tag is the single version source (`-ldflags -X`, `-trimpath`, `CGO_ENABLED=0`, fallback to `debug.ReadBuildInfo`); delete `version.txt` (`release.yaml` tags from it on every push to main today)
4. Release workflow: build matrix (darwin/linux x amd64/arm64), `sha256sum`, `gh release create`
5. README: fix Go "1.16+" (go.mod requires 1.26), document tmux 3.8+, install, completion setup, exit codes, JSON format, an fzf recipe, how it differs from tmux-resurrect, tmuxinator, sesh

## Tests

I write every test myself, by hand, to learn how they work. Claude only reviews them. Written so far: `utils` (`Quote`, `Escape`, `Slug`). `cli`, `cmd`, `config` and `tmux` have no tests. The items below include areas that used to be covered (`config.New`, `cli` help and lookup, `utils.Confirm`, `Line` conversions, pane builders, `buildNewSessionCmd`, `newSessionError`, window layout normalization and `buildSetLayoutCmd`).

### Seam and coverage

1. `fakeRunner` recording args and returning canned output
2. Tests for `CreateSessions`, `CreateWindows`, `Snapshot`, `ListSessions` through the fake
3. Command-level tests with injected stdin/stdout/stderr (none exist yet; most need the runner)
4. Add a `Run` test for each command's success path once a fake runner exists
5. Tests that flags reach their fields (`-json`, `-path`, `-menu`): a value receiver or a nil `Flags()` broke this silently before
6. Use `t.Helper`, `t.Cleanup`, `t.Setenv` in tests (`just test` already runs with `-race`)

### Builders and data model

1. Builder tests (`buildNewWindowCmd`, `buildSplitPaneCmd`, `buildRenamePaneCmd`, `buildSetLayoutCmd`): session `a#b` keeps a single `#` in `-t`, title `#S` becomes `##S` after `-T`; session names and the tmux output format are not covered yet
2. `buildNewWindowsCmd` argv table: `";"` between commands and none at the ends, `-c` omitted for an empty path, a window name with `#`, a missing panes key and an empty panes slice both return `ErrWindowWithoutPanes`; `buildSetLayoutCmd` passes the raw layout text unchanged; a window with no layout gets no `select-layout`; `appendCmd`
3. Round trip: `DesanitizeTemplSnapshot(SanitizeSnapshot(x, n), n)` restores the keys and names for slug-safe and unsafe names (`my proj`, `a.b`)
4. Template names through the commands: `new-session -templ "my proj"` finds `my-proj.json` (`HasTemplate` and `ShowTemplate` agree), `save-template` and `delete-template` use the same file, a name that slugs to `""` is rejected
5. Error paths: `SanitizeSnapshot` with an unknown session name; `DesanitizeTemplSnapshot` with a missing `Sessions`, `Windows` or `Panes` key; check `errors.Is` with the sentinel
6. `SnapshotData.String` tree: last-child prefixes (`├─`, `└─`, rails), name column alignment (including multibyte names), empty pane path, several sessions separated by a blank line, windows ordered by index, a loaded snapshot (empty `Window.Slug`) still lists its panes
7. `printJSON`, `list-templates -json` (empty list prints `[]`, not `null`), `version -json`, `paths` and `paths -json` (three lines in the order config, templates, sessions; creates nothing)
8. `Info()` of the six commands that format `DescLong` from `cfg` (`%s` is filled with the real path, no stray `%!s(MISSING)`)
9. `parseTmuxLines`: ordering, a duplicate session (hit keeps the stored session and stores it again) and a duplicate window ID (one window entry)
10. `utils.Confirm`, `Line` conversions, pane builders, `buildNewSessionCmd`, `newSessionError`
11. Window layout: decoding real tmux text, `isSplitLayout` (single pane, split, empty, invalid), `Data.String` and `Data.JSON` output, `buildSetLayoutCmd`, `select-layout` after the splits

### Config and CLI

1. `config.New` (`t.Setenv` for `XDG_CONFIG_HOME` and `HOME`, version trimming, missing home)
2. Missing config dir (`t.TempDir()` plus `t.Setenv("XDG_CONFIG_HOME", ...)`): `config.New` and the read-only calls (`ListTemplates` returns an empty list, `HasTemplate` returns `(false, nil)`, `ShowTemplate` and `DeleteTemplate` return an error) create no directory; `SaveSessions` and `SaveTemplate` create it
3. `Config.SaveSessions` with `t.TempDir()`: two consecutive calls (dir missing, then dir exists), read the file back
4. Config store round trip, missing file, bad JSON, traversal name (rejected by `os.Root`), non-`.json` file, delete missing, using `t.TempDir()` (save, show and delete all exist now; no tests yet)
5. `internal/cli` help output (order, aliases), command lookup, `ErrUnknownCommand`, `-h` handling
6. `Cli.Run` adds `\nUsage: ...` to an `ErrUsage` error and keeps `errors.Is(err, cmd.ErrUsage)` true; `Info.Usage()` with and without `Arg`; `new-session` on a duplicate session returns an error that satisfies `errors.Is(err, tmux.ErrSessionExists)` (needs the runner seam)

### Manual verification against tmux

1. `select-layout` in the argv chain on a multi-pane template (the layout is passed raw as one argv element; the `ble` template has no layout, so this is untried; tmux needs the pane count to match)
2. Whether `-c` (start directory) is format-expanded; escape `#` in paths if it is
3. A window name or pane title ending in `;`: tmux treats a trailing `;` on an argv word as a command separator, so it may need `\;` even in argv form
4. On a throwaway server (private socket, `-L`): single-string chain parsing, `'it'\''s'` quoting, `new-window -k` on a session's only window, `-T` format expansion, order of `list-panes -a`, session names with `.` or `:`
   - Verified (default server, not a private socket): `new-session -s` is format-expanded (`-s 'a##b'` creates `a#b`); `-t` is not (`has-session -t '=a#b'` finds it, `'=a##b'` does not)
   - Verified: duplicate `new-session` exits 1 with `duplicate session: NAME` and stops a chain; `new-session -A -d` on an existing session fails without a terminal; `-c` with a missing directory exits 0 and falls back to `$HOME`
   - Verified: tmux 3.8 prints `#{window_layout}` as JSON; it can contain keys the structs do not model (`"l":0`), so the raw text is what gets sent back to `select-layout`

## Nice to have

Not needed for 0.8.0. Pick up when the core is solid.

### CLI polish

1. Unknown command: "did you mean"
2. `ls` / `list-templates`: plain output when piped, `-0` (`-json` is done)
3. `save -o -`, `restore -f FILE`, `show-template NAME`, `attach -` (name from stdin)
4. TTY detection; `NO_COLOR` and `TERM=dumb` respected if color is added
5. Handle SIGPIPE / EPIPE on closed stdout (`tmxu ls | head -1`)
6. `//go:build !windows` or a clear unsupported-OS message

### Menu (`attach -menu`)

1. `attach -menu`: implement it (the flag stays; help says "not implemented yet"); drop the `// TODO: implement fzf style menu` comment in `Run` when done
2. Menu: `term.IsTerminal` check with a plain-list fallback; handle `MakeRaw` and read errors, EOF, Ctrl-C (byte 3), empty items, titles longer than the pad width; multibyte-safe padding
3. Menu: draw on stderr or `/dev/tty`; alternate screen; escapes as constants; restore the terminal on SIGTERM/SIGHUP through context cancellation

### Performance (parked)

Symptom: after restore, panes "catch up" (shells still booting, prompts redraw) when the user attaches. Baseline first: `time tmxu restore`. If under about 2 s, the first two items are probably enough.

1. `new-session -x <cols> -y <rows>` from `term.GetSize`; no flags when stdout is not a TTY
2. Wait for panes to be ready: poll `capture-pane -p -t <pane>` every ~50 ms, ~5 s timeout per session, sessions in parallel; `waiting for shells...` on stderr; warn on timeout; `-no-wait` to skip
3. Batch a window's tmux calls into one invocation (low priority)
4. Staggered start in batches of 3-4 sessions (only if still janky)
5. Parallel snapshot with `sync.WaitGroup`, indexed result slice and a semaphore (measure first); restore stays sequential

### Distribution and docs

1. `tmxu completion zsh|bash|fish` (hand-written, `//go:embed` or generated from the registry); hidden `__complete` for session and template names; tests for it
2. Man page `man/tmxu.1` (NAME, SYNOPSIS, COMMANDS, OPTIONS, FILES, ENVIRONMENT, EXIT STATUS, EXAMPLES, SEE ALSO)
3. Homebrew tap or formula, installing the man page and completions (or remove references)
4. Pin GitHub Actions by SHA; add Dependabot for modules and actions
5. `CONTRIBUTING.md`, `CHANGELOG.md`, `.editorconfig`
6. Golden files in `testdata/` with an `-update` flag (saved JSON, help output, `ls` table)
7. Optional integration test on a private socket (`tmux -L tmxu-test-<name>`), skipped without tmux

## Decisions

Settled so far: window layouts are stored as raw tmux JSON and applied with `select-layout` after the splits (the normalized `LayoutNode` code is deleted), errors, prompts and diagnostics go to stderr, data and success messages to stdout, a window without panes is an error (`ErrWindowWithoutPanes`, no `panic`), one argv chain with `";"` elements (no `source-file`), derived pane names `<window>_<index>`, on-disk format reuses the `internal/tmux` structs, `dump -session` prints real data (sanitizing lives in `save-template`), text output is a names-only tree (no ids, indexes or layouts; those stay in `-json`), JSON shapes are `{"sessions":[...]}`, `{"templates":[...]}`, `{"version":"..."}` and `{"config","templates","sessions"}` (paths), and flags have one name each (`-yes`, no short alias; `flag` has none).

1. Existing sessions on restore: skip, refuse, or force
2. Wait for shells by default with `-no-wait`, or opt-in `-wait`
3. `save-template` when the template exists: refuse, overwrite, or require `-f`
4. Template identity is the slug on disk: `My Proj` and `my-proj` are the same file (silent overwrite), and `list-templates` shows slugs, not the names the user typed; accept that, or store the display name inside the file
5. Template start directories: `SanitizeSnapshot` clears every `Pane.Path`, so `new-session -templ` windows start in tmux's default directory and `-path` is ignored for them; keep that, apply `-path` to every window and pane, or keep paths in templates
