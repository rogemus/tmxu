# TODO

Everything left to do to finish porting `restore` into `internal/` and to ship `tmxu` as a good Unix tool.

Policy: stdlib-first. Only `golang.org/x/term` is allowed as a dependency (currently not in `go.mod`; add it with the menu or `term.GetSize` work). Write things by hand to learn Go. I write every test myself, by hand; Claude only reviews them.
Scope: single user, so there is no migration from the old `tmux-sessions.json`. The new schema is designed clean.
Target: tmux 3.8+ only (JSON `window_layout`).

Status: `main` is wired to `internal/cli`. `Run` dispatches `ls`, `save`, `version`, `help`, `attach`, `dump`, `new-session`, `restore` (registered, `Run` is still a stub), `list-templates`, `save-template`, `delete-template`, `show-template`, `paths`. Config file operations go through `os.Root` (helpers take a `*os.Root` and a root-relative name, and return `filepath.Join(root.Name(), name)`); this covers file reads, writes and deletes only, directory creation and `ListTemplates` still use plain paths. Each command owns its `FlagSet` (built in the constructor, bound to its own fields, pointer receivers); `Cmd.Flags()` is the single source of flags for parsing and help. Restore and the menu are not ported. `new-session -templ` works end to end (error paths are still open). `dump` and `show-template` print a names-only tree (session, windows, panes with their paths); `-json` works on `ls`, `list-templates`, `version`, `paths`, `dump` and `show-template`; `save` has `-yes`. `save` writes `tmux-sessions2.json` into the config dir (first version: not atomic, files 0600, dirs 0700). Window layouts are read, stored as raw tmux JSON and applied after the splits. The persisted shape is the nested `SnapshotData` (version 1). `internal/` tests: only `utils` (`Quote`, `Escape`, `Slug`) plus the `internal/expect` helper, all written by hand. Everything that shells out to tmux needs the runner seam.

Done items are removed from this file; `git log` has them. Goal: 0.8.0 release. Port `restore` to `internal/` (the old `cli/` package is already deleted). `attach -menu` is not ported yet: it was broken in the previous version. The flag stays registered (and `restore -force`) and will be implemented.

Every item starts with a type and a difficulty tag, for example `1. [BUG] #HARD ...`.
- Type: `[BUG]`, `[FEATURE]`, `[REFACTOR]`, `[TEST]`, `[DECISION]`, `[VERIFY]`, `[RELEASE]`, `[DOCS]`, `[COSMETIC]`.
- Difficulty: `#EASY` is one small change in one place, `#MEDIUM` touches a few files or needs a small design choice, `#HARD` is a design question or a cross-cutting change.
- Numbers restart in each section and shift when an item is removed.

Sections: [Must Fix](#must-fix) · [Should Fix](#should-fix) · [Nice to have](#nice-to-have) · [Cosmetics](#cosmetics) · [Settled decisions](#settled-decisions)

## Must Fix

Needed before 0.8.0: data loss, a broken `save`, the `restore` port itself, decisions the code is waiting on, and release mechanics.

1. [BUG] #HARD no rollback: a failed window step leaves a half-built session that blocks the retry with `duplicate session`; applies to `new-session -templ` and to restore (roll back sessions created in this run when a later step fails)
2. [BUG] #MEDIUM `CreateSessions` and `buildNewSessionsCmd` (`session.go`) have no callers and still pass one joined string to `exec.Command`; rewrite on argv for restore (or delete them)
3. [FEATURE] #HARD `restore-sessions` (`restore`, `r`), one shared `restoreSession` for restore and new-from-template
   - `-force`
4. [FEATURE] #MEDIUM check `has-session` before creating; skip or refuse existing sessions (today only `NewSession` detects a duplicate, from tmux's stderr; a duplicate in a `CreateSessions` chain stops the rest)
5. [FEATURE] #MEDIUM never use `-k` on a session that was not created in this run
6. [FEATURE] #MEDIUM explicit force path with a prompt that lists what will be killed; refuse to kill the current session (`$TMUX`, `display-message -p '#S'`)
7. [FEATURE] #MEDIUM read the real `base-index` and `pane-base-index` (restore targets `=S:W.<saved pane index>`, so it only works when the restoring server uses the same bases as the saving one; `BaseIndex` was deleted as unused; `tmux show-options -gv base-index` reads the global option, the session's own may differ); handle saved index != current base so no blank leftover window stays
8. [FEATURE] #HARD decide split direction and the 80x24 detached-size problem (`new-session -x -y`, or apply layout per split)
9. [FEATURE] #MEDIUM config store `writeFileAtomic` (temp file, `Sync`, `Rename`, same `0600` mode), keep a `.bak`
10. [DECISION] #EASY existing sessions on restore: skip, refuse, or force
11. [DECISION] #EASY `save-template` when the template exists: refuse, overwrite, or require `-f` (it truncates silently today)
12. [DECISION] #EASY template start directories: `AsTemplate` clears every pane path, so `new-session -templ` windows start in tmux's default directory and `-path` is ignored for them; keep that, apply `-path` to every window and pane, or keep paths in templates
13. [VERIFY] #EASY `select-layout` in the argv chain on a multi-pane template (`ran_liftup`, `dt_demo_now` window 2): the layout is raw tmux JSON that still carries the old pane ids (`"I":"%5"`) and `"a":true`; check that tmux accepts it for the new panes, and if not strip `I` in `buildSetLayoutCmd`; tmux needs the pane count to match
14. [RELEASE] #EASY switch `save` back from `tmux-sessions2.json` to `tmux-sessions.json` (temporary name while the port is in progress); change only `sessionsFileName` (`paths` and the `DescLong` text follow it) and update the status text
15. [RELEASE] #MEDIUM release on `v*` tag push; git tag is the single version source (`-ldflags -X`, `-trimpath`, `CGO_ENABLED=0`, fallback to `debug.ReadBuildInfo`); delete `version.txt` (`release.yaml` tags from it on every push to main today)
16. [RELEASE] #MEDIUM workflow build matrix (darwin/linux x amd64/arm64), `sha256sum`, `gh release create`
17. [DOCS] #MEDIUM README: fix Go "1.16+" (go.mod requires 1.26), document tmux 3.8+, install, completion setup, exit codes, JSON format, an fzf recipe, how it differs from tmux-resurrect, tmuxinator, sesh

## Should Fix

Not blocking, but restore is risky or unpleasant without them. Roughly in the order to do them: runner seam and tests first, then errors and signals, then CLI behavior.

1. [REFACTOR] #HARD one `run()` helper or a `Runner` interface (`Run`, `Attach`) with an `execRunner`; no direct `exec.Command` elsewhere
2. [REFACTOR] #MEDIUM commands take their tmux dependency through an interface instead of calling package functions
3. [REFACTOR] #MEDIUM typed `*TmuxError{Args, Stderr, Err}`; include stderr in messages
4. [REFACTOR] #EASY sentinel `ErrNoTmux` (`exec.ErrNotFound`); "tmux not found in PATH" message (`ErrNoSession` and `ErrSessionExists` live in `tmux`, `ErrNoTemplate` in `cmd`)
5. [REFACTOR] #EASY `errors.Join` to collect restore errors
6. [FEATURE] #EASY "No server running" handled like `tmux ls`
7. [FEATURE] #EASY `exec.LookPath("tmux")` and `tmux -V` check at startup with a clear message
8. [FEATURE] #HARD thread `ctx`; `signal.NotifyContext` in `main`; `exec.CommandContext` with a 5s timeout for non-interactive calls, none for `attach`
9. [FEATURE] #EASY exit code 130 on interrupt (0, 1, 2 done)
10. [FEATURE] #MEDIUM Ctrl-C during restore: clean up or report the partial session
11. [FEATURE] #MEDIUM restore the original active pane and window
12. [FEATURE] #EASY `attach` / `switch-client` when `$TMUX` is set; `=name` exact target
13. [FEATURE] #EASY start directory guard (old `startDirArgs`): verified that tmux does not fail on a missing `-c` directory, it silently starts in `$HOME`; decide whether to warn
14. [FEATURE] #EASY templates: clear missing-file error for `show-template` / `delete-template`, and `-f` for `delete-template`
15. [FEATURE] #MEDIUM config: `TMXU_CONFIG_DIR` env and `--config-dir`; ignore a relative `XDG_CONFIG_HOME`; doc comments on `Config`, `BaseDirPath`, `AppDirPath`, `New`
16. [FEATURE] #HARD `-yes` for restore, new and delete-template (done for `save-sessions`), `-n` / `--dry-run` (restore, new, delete-template, save), `-q` / `-v` with `log/slog` to stderr (`-v` logs tmux args)
17. [FEATURE] #MEDIUM help layout: Usage, Commands, Options, Environment, Exit status
18. [DECISION] #EASY what bare `tmxu` does (help + exit 0 today)
19. [DECISION] #EASY wait for shells by default with `-no-wait`, or opt-in `-wait`
20. [DECISION] #EASY template identity is the slug on disk: `My Proj` and `my-proj` are the same file (silent overwrite), and `list-templates` shows slugs, not the names the user typed; accept that, or store the display name inside the file
21. [DECISION] #EASY a pane path containing `\n` splits one pane into two rows, so `parseLines` fails with `missing parts` and `save` of that session errors; accept and document it, skip the pane with a warning, or keep failing loudly
22. [BUG] #MEDIUM `--path` is not applied to template windows or panes (`AsTemplate` clears every pane path and the `-c` flag is skipped for an empty path); follows the template start directories decision in Must Fix
23. [BUG] #MEDIUM session names with `.` or `:` (tmux may rewrite them, so the printed name and `attach` could disagree; verify, then reject, print the real name, or target by session id)
24. [BUG] #MEDIUM flags after positional args work; `--` terminator honored (today `-path` after a name is silently ignored)
25. [BUG] #EASY only `new-session` rejects extra positional args (`ErrUnexpectedArgument`); `dump`, `ls`, `version`, `paths`, `save`, `attach`, `save-template`, `show-template` and `delete-template` ignore them
26. [BUG] #EASY `utils.Confirm`: one reader, handle the read error, fail loudly when stdin is not a TTY (becomes a Must Fix once the `-force` prompt ships)
27. [TEST] #MEDIUM `fakeRunner` recording args and returning canned output
28. [TEST] #MEDIUM `CreateSessions`, `CreateWindows`, `Snapshot`, `ListSessions` through the fake
29. [TEST] #MEDIUM command-level tests with injected stdin/stdout/stderr, and a `Run` success path for each command (none exist yet; most need the runner)
30. [TEST] #EASY builders (`buildNewWindowCmd`, `buildSplitPaneCmd`, `buildRenamePaneCmd`, `buildSetLayoutCmd`): session `a#b` keeps a single `#` in `-t`, title `#S` becomes `##S` after `-T`; session names and the tmux output format are not covered yet
31. [TEST] #MEDIUM `buildNewWindowsCmd` argv table: `";"` between commands and none at the ends, `-c` omitted for an empty path, a window name with `#`, a missing panes key and an empty panes slice both return `ErrWindowWithoutPanes`; `buildSetLayoutCmd` passes the raw layout text unchanged; a window with no layout gets no `select-layout`; `appendCmd`
32. [TEST] #EASY `parseTmuxLines`: one window with one pane; two windows with one pane each (regression: every pane once landed in window 0); one window with two panes keeps pane order; two sessions that both have window 1 stay apart; `Layout` is kept only for split windows (`isSplitLayout`)
33. [TEST] #MEDIUM `parseLines` and `parseSessionLines`: names and paths with `"`, `\`, `#` and spaces come back unchanged; a row with too few fields returns `ErrCannotParseLine`; a non-numeric window or pane index returns an error that satisfies `errors.Is`; the last field keeps a stray `\x1f` (`SplitN`); empty output; `HasSession` for exit 0, exit 1 and any other failure (needs the runner seam); `Snapshot("nosuch")` returns `ErrNoSession`
34. [TEST] #MEDIUM template flow: `AsTemplate` clears the session name and every pane path and keeps windows, indexes and layouts; `Template()` returns `ErrInvalidTemplate` for 0 and for 2 sessions; `RestoreFromTempl` validates before it touches tmux (needs the runner seam); `show-template` on a file with no session returns the wrapped error
35. [TEST] #MEDIUM template names through the commands: `new-session -templ "my proj"` finds `my-proj.json` (`HasTemplate` and `ShowTemplate` agree), `save-template` and `delete-template` use the same file, a name that slugs to `""` is rejected
36. [TEST] #MEDIUM missing config dir (`t.TempDir()` plus `t.Setenv("XDG_CONFIG_HOME", ...)`): `config.New` and the read-only calls (`ListTemplates` returns an empty list, `HasTemplate` returns `(false, nil)`, `ShowTemplate` and `DeleteTemplate` return an error) create no directory; `SaveSessions` and `SaveTemplate` create it
37. [TEST] #MEDIUM config store round trip, missing file, bad JSON, traversal name (rejected by `os.Root`), non-`.json` file, delete missing, using `t.TempDir()`; `Config.SaveSessions` two consecutive calls (dir missing, then dir exists), read the file back
38. [VERIFY] #EASY whether `-c` (start directory) is format-expanded; escape `#` in paths if it is
39. [VERIFY] #EASY a window name or pane title ending in `;`: tmux treats a trailing `;` on an argv word as a command separator, so it may need `\;` even in argv form
40. [VERIFY] #EASY the `\x1f` field separator (`lineSafeSeparator`): check on a throwaway server that tmux prints it raw (`list-sessions -F $'#{session_name}\x1f#{session_id}' | od -c`) and what it does to a `\n` or control character in a session or window name; `parseLines` assumes neither reaches the output
41. [VERIFY] #MEDIUM on a throwaway server (private socket, `-L`): single-string chain parsing, `'it'\''s'` quoting, `new-window -k` on a session's only window, `-T` format expansion, order of `list-panes -a`, session names with `.` or `:`
    - Verified (default server, not a private socket): `new-session -s` is format-expanded (`-s 'a##b'` creates `a#b`); `-t` is not (`has-session -t '=a#b'` finds it, `'=a##b'` does not)
    - Verified: duplicate `new-session` exits 1 with `duplicate session: NAME` and stops a chain; `new-session -A -d` on an existing session fails without a terminal; `-c` with a missing directory exits 0 and falls back to `$HOME`
    - Verified: tmux 3.8 prints `#{window_layout}` as JSON; it can contain keys the structs do not model (`"l":0`), so the raw text is what gets sent back to `select-layout`

## Nice to have

Not needed for 0.8.0. Pick up when the core is solid.

1. [FEATURE] #EASY unknown command: "did you mean"
2. [FEATURE] #EASY `ls` / `list-templates`: plain output when piped, `-0` (`-json` is done)
3. [FEATURE] #MEDIUM `save -o -`, `restore -f FILE`, `show-template NAME`, `attach -` (name from stdin)
4. [FEATURE] #EASY TTY detection; `NO_COLOR` and `TERM=dumb` respected if color is added
5. [FEATURE] #EASY handle SIGPIPE / EPIPE on closed stdout (`tmxu ls | head -1`)
6. [FEATURE] #EASY `//go:build !windows` or a clear unsupported-OS message
7. [FEATURE] #HARD `attach -menu`: implement it (the flag stays; help says "not implemented yet"); drop the `// TODO: implement fzf style menu` comment in `Run` when done
8. [FEATURE] #HARD menu: `term.IsTerminal` check with a plain-list fallback; handle `MakeRaw` and read errors, EOF, Ctrl-C (byte 3), empty items, titles longer than the pad width; multibyte-safe padding
9. [FEATURE] #MEDIUM menu: draw on stderr or `/dev/tty`; alternate screen; escapes as constants; restore the terminal on SIGTERM/SIGHUP through context cancellation
10. [FEATURE] #MEDIUM `tmxu completion zsh|bash|fish` (hand-written, `//go:embed` or generated from the registry); hidden `__complete` for session and template names; tests for it
11. [DOCS] #MEDIUM man page `man/tmxu.1` (NAME, SYNOPSIS, COMMANDS, OPTIONS, FILES, ENVIRONMENT, EXIT STATUS, EXAMPLES, SEE ALSO)
12. [RELEASE] #MEDIUM Homebrew tap or formula, installing the man page and completions (or remove references)
13. [RELEASE] #EASY CI: `main-check.yaml` already runs build, vet, gofmt and `go test -race ./...`; add staticcheck and a `go mod tidy` diff
14. [RELEASE] #EASY pin GitHub Actions by SHA; add Dependabot for modules and actions
15. [DOCS] #EASY `CONTRIBUTING.md`, `CHANGELOG.md`, `.editorconfig`
16. [TEST] #MEDIUM golden files in `testdata/` with an `-update` flag (saved JSON, help output, `ls` table)
17. [TEST] #MEDIUM optional integration test on a private socket (`tmux -L tmxu-test-<name>`), skipped without tmux
18. [TEST] #EASY tests that flags reach their fields (`-json`, `-path`, `-menu`): a value receiver or a nil `Flags()` broke this silently before
19. [TEST] #MEDIUM `SnapshotData.String` tree: last-child prefixes (`├─`, `└─`, rails), name column alignment (including multibyte names), empty pane path, several sessions separated by a blank line, windows ordered by index
20. [TEST] #EASY `printJSON`, `list-templates -json` (empty list prints `[]`, not `null`), `version -json`, `paths` and `paths -json` (three lines in the order config, templates, sessions; creates nothing)
21. [TEST] #EASY `Info()` of the six commands that format `DescLong` from `cfg` (`%s` is filled with the real path, no stray `%!s(MISSING)`)
22. [TEST] #MEDIUM `utils.Confirm`, `Line` conversions, pane builders, `buildNewSessionCmd`, `newSessionError`
23. [TEST] #MEDIUM window layout: decoding real tmux text, `isSplitLayout` (single pane, split, empty, invalid), `Data.String` and `Data.JSON` output, `buildSetLayoutCmd`, `select-layout` after the splits
24. [TEST] #EASY `config.New` (`t.Setenv` for `XDG_CONFIG_HOME` and `HOME`, version trimming, missing home)
25. [TEST] #EASY `internal/cli` help output (order, aliases), command lookup, `ErrUnknownCommand`, `-h` handling
26. [TEST] #MEDIUM `Cli.Run` adds `\nUsage: ...` to an `ErrUsage` error and keeps `errors.Is(err, cmd.ErrUsage)` true; `Info.Usage()` with and without `Arg`; `new-session` on a duplicate session returns an error that satisfies `errors.Is(err, tmux.ErrSessionExists)` (needs the runner seam)
27. [TEST] #EASY use `t.Helper`, `t.Cleanup`, `t.Setenv` in tests (`just test` already runs with `-race`)
28. [FEATURE] #HARD performance (parked): symptom is panes "catching up" (shells still booting, prompts redraw) after restore; baseline first with `time tmxu restore`, and if it is under about 2 s the first two items are probably enough
    - `new-session -x <cols> -y <rows>` from `term.GetSize`; no flags when stdout is not a TTY
    - Wait for panes to be ready: poll `capture-pane -p -t <pane>` every ~50 ms, ~5 s timeout per session, sessions in parallel; `waiting for shells...` on stderr; warn on timeout; `-no-wait` to skip
    - Batch a window's tmux calls into one invocation (low priority)
    - Staggered start in batches of 3-4 sessions (only if still janky)
    - Parallel snapshot with `sync.WaitGroup`, indexed result slice and a semaphore (measure first); restore stays sequential

## Cosmetics

Naming, text, structure, duplication, doc comments. No behavior change for users.

1. [COSMETIC] #EASY `helpCmd`: align the FLAGS descriptions in a column (`-path` and `-templ` differ in length, the template pads with fixed spaces); use `text/tabwriter` like `listAllCommands`, with a tab between name and description, flushed after `Execute`; check that the other sections are not affected
2. [COSMETIC] #MEDIUM doc comments on exported identifiers (package docs exist everywhere; `Info`, `Flag`, `IOs`, `Cmd`, `Config`, `New`, `SnapshotData`, `SessionData`, `WindowData`, `PaneData`, constructors and `utils.Slug` have none); unexport what is not API
3. [COSMETIC] #EASY `utils.Quote` has no callers on the argv path; only `buildNewSessionsCmd` (dead, `GetName` is gone) uses it (`printError` is gone); delete it with `CreateSessions` and `buildNewSessionsCmd`. Note `new-session -s` is format-expanded, so a rewrite on argv needs `utils.Escape` for the name
4. [COSMETIC] #EASY `HasTemplate` duplicates what `ShowTemplate` reports; keep it, or return a not-found sentinel from `ShowTemplate`
5. [COSMETIC] #EASY config: fold the repeated `os.OpenRoot` + `defer root.Close()` (5 places: four in `templates.go`, one in `sessions.go`) into one helper
6. [COSMETIC] #MEDIUM generic `readJSON[T]` / `writeJSON`; `os.MkdirAll` (file reads, writes and deletes already go through `os.Root`: `readFile`, `saveToFile`, `deleteFile` in `tmxu_config.go`)
7. [COSMETIC] #EASY `Root`-based listing for `ListTemplates` (`fs.ReadDir(root.FS(), ".")`; `os.Root` has no `ReadDir`; `readTemplatesName` then takes `[]fs.DirEntry`)

## Settled decisions

Window layouts are stored as raw tmux JSON and applied with `select-layout` after the splits (the normalized `LayoutNode` code is deleted). Errors, prompts and diagnostics go to stderr, data and success messages to stdout. A window without panes is an error (`ErrWindowWithoutPanes`, no `panic`). One argv chain with `";"` elements (no `source-file`). Derived pane names `<window>_<index>`. On-disk format is the nested `SnapshotData`: `{"version":1,"sessions":[{"session_name","session_windows":[{"window_name","window_index","window_layout"?,"window_panes":[{"pane_index","pane_path"}]}]}]}`; no slugs, no tmux ids, no version check and no migration (never released). A template is a snapshot with exactly one session whose name and pane paths are cleared (`AsTemplate`); `SnapshotData.Template` validates that and `RestoreFromTempl` builds the session from it. `dump -session` prints real data (sanitizing lives in `save-template`). Text output is a names-only tree (no ids, indexes or layouts; those stay in `-json`). JSON shapes are `{"sessions":[...]}`, `{"templates":[...]}`, `{"version":"..."}` and `{"config","templates","sessions"}` (paths). Flags have one name each (`-yes`, no short alias; `flag` has none).
