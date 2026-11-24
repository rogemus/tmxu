# TODO

Everything left to do to finish porting `restore` into `internal/` and to ship `tmxu` as a good Unix tool.

Policy: stdlib-first. Only `golang.org/x/term` is allowed as a dependency (in `go.mod`; used by `utils.RequireTerminal`, the menu and `term.GetSize` work will reuse it). Write things by hand to learn Go. I write every test myself, by hand; Claude only reviews them.
Scope: single user, so there is no migration from the old `tmux-sessions.json`. The new schema is designed clean.
Target: tmux 3.8+ only (JSON `window_layout`).

Status: `main` is wired to `internal/cli`. `Run` dispatches `ls`, `save`, `version`, `help`, `attach`, `dump`, `new-session`, `restore` (works: refuses sessions that already exist, `-force` kills them first and the prompt lists the sessions it will kill, `-yes` skips the prompt; killing the current session is allowed on purpose, so there is no guard; `ListSessions` returns an empty list when the server is not running; rollback is deferred to 0.9.0), `list-templates`, `save-template`, `delete-template`, `show-template`, `paths`. Config file operations go through `os.Root` (helpers take a `*os.Root` and a root-relative name, and return `filepath.Join(root.Name(), name)`); this covers file reads, writes and deletes only, directory creation and `ListTemplates` still use plain paths. Each command owns its `FlagSet` (built in the constructor, bound to its own fields, pointer receivers); `Cmd.Flags()` is the single source of flags for parsing and help. The menu is not ported. `new-session -templ` works end to end (error paths are still open). `dump` and `show-template` print a names-only tree (session, windows, panes with their paths); `-json` works on `ls`, `list-templates`, `version`, `paths`, `dump` and `show-template`; `save` has `-yes`. `save` writes `tmux-sessions.json` into the config dir (files 0600, dirs 0700). Writes are atomic: temp file `<name>.tmp`, `Sync`, `Close`, a hard link `<name>.bak` of the old file, `Rename` over the target, then a directory `Sync`; so `<name>` exists at every instant and `<name>.bak` holds the previous version (templates use the same helper, `ListTemplates` ignores the extra files by extension). Window layouts are read, stored as raw tmux JSON and applied after the splits. The persisted shape is the nested `SnapshotData` (version 1). `internal/` tests: only `utils` (`Escape`, `Slug`) plus the `internal/expect` helper, all written by hand. Everything that shells out to tmux needs the runner seam.

Done items are removed from this file; `git log` has them. Goal: 0.8.0 release. Port `restore` to `internal/` (the old `cli/` package is already deleted). `attach -menu` is not ported yet: it was broken in the previous version. The flag stays registered (and `restore -force`) and will be implemented.

Every item starts with a type and a difficulty tag, for example `1. [BUG] #HARD ...`.

- Type: `[BUG]`, `[FEATURE]`, `[REFACTOR]`, `[TEST]`, `[DECISION]`, `[VERIFY]`, `[RELEASE]`, `[DOCS]`, `[COSMETIC]`.
- Difficulty: `#EASY` is one small change in one place, `#MEDIUM` touches a few files or needs a small design choice, `#HARD` is a design question or a cross-cutting change.
- Numbers restart in each section and shift when an item is removed.

Sections: [Must Fix](#must-fix) · [Should Fix](#should-fix) · [Verify](#verify) · [Nice to have](#nice-to-have) · [0.9.0 plans](#090-plans) · [Cosmetics](#cosmetics) · [Settled decisions](#settled-decisions)

## Must Fix

Needed before 0.8.0: data loss, a broken `save`, the `restore` port itself, decisions the code is waiting on, and release mechanics. The checks that need a real run, including the ones that block the release, are in [Verify](#verify).

Nothing left here. The remaining 0.8.0 blockers are in [Verify](#verify).

## Should Fix

Not blocking, but restore is risky or unpleasant without them. Roughly in the order to do them: runner seam and tests first, then errors and signals, then CLI behavior.

1. [REFACTOR] #HARD one `run()` helper or a `Runner` interface (`Run`, `Attach`) with an `execRunner`; no direct `exec.Command` elsewhere
2. [REFACTOR] #MEDIUM commands take their tmux dependency through an interface instead of calling package functions
3. [REFACTOR] #MEDIUM typed `*TmuxError{Args, Stderr, Err}`; include stderr in messages
4. [REFACTOR] #EASY sentinel `ErrNoTmux` (`exec.ErrNotFound`); "tmux not found in PATH" message (`ErrNoSession` and `ErrSessionExists` live in `tmux`, `ErrNoTemplate` in `cmd`)
5. [REFACTOR] #EASY `errors.Join` to collect restore errors
6. [FEATURE] #EASY `exec.LookPath("tmux")` and `tmux -V` check at startup with a clear message
7. [FEATURE] #HARD thread `ctx`; `signal.NotifyContext` in `main`; `exec.CommandContext` with a 5s timeout for non-interactive calls, none for `attach`
8. [FEATURE] #EASY exit code 130 on interrupt (0, 1, 2 done)
9. [FEATURE] #MEDIUM Ctrl-C during restore: clean up or report the partial session
10. [FEATURE] #MEDIUM restore the original active pane and window
11. [FEATURE] #EASY `attach` / `switch-client` when `$TMUX` is set; `=name` exact target
12. [FEATURE] #EASY start directory guard (old `startDirArgs`): verified that tmux does not fail on a missing `-c` directory, it silently starts in `$HOME`; decide whether to warn
13. [FEATURE] #EASY templates: clear missing-file error for `show-template` / `delete-template`, and `-f` for `delete-template`
14. [FEATURE] #MEDIUM config: `TMXU_CONFIG_DIR` env and `--config-dir`; ignore a relative `XDG_CONFIG_HOME`; doc comments on `Config`, `BaseDirPath`, `AppDirPath`, `New`
15. [FEATURE] #HARD `-yes` for new and delete-template (done for `save-sessions` and `restore`), `-n` / `--dry-run` (restore, new, delete-template, save), `-q` / `-v` with `log/slog` to stderr (`-v` logs tmux args)
16. [FEATURE] #MEDIUM help layout: Usage, Commands, Options, Environment, Exit status
17. [DECISION] #EASY what bare `tmxu` does (help + exit 0 today)
18. [DECISION] #EASY wait for shells by default with `-no-wait`, or opt-in `-wait`
19. [DECISION] #EASY template identity is the slug on disk: `My Proj` and `my-proj` are the same file (silent overwrite), and `list-templates` shows slugs, not the names the user typed; accept that, or store the display name inside the file
20. [DECISION] #EASY a pane path containing `\n` splits one pane into two rows, so `parseLines` fails with `missing parts` and `save` of that session errors; accept and document it, skip the pane with a warning, or keep failing loudly
21. [BUG] #MEDIUM session names with `.` or `:` (tmux may rewrite them, so the printed name and `attach` could disagree; verify, then reject, print the real name, or target by session id)
22. [BUG] #MEDIUM flags after positional args work; `--` terminator honored (today `-path` after a name is silently ignored)
23. [BUG] #EASY only `new-session` rejects extra positional args (`ErrUnexpectedArgument`); `dump`, `ls`, `version`, `paths`, `save`, `attach`, `save-template`, `show-template` and `delete-template` ignore them
24. [TEST] #MEDIUM `fakeRunner` recording args and returning canned output
25. [TEST] #MEDIUM `CreateSessions`, `CreateWindows`, `Snapshot`, `ListSessions` through the fake; `ListSessions`: exit 1 with `no server` on stderr returns `[]SessionData{}` and nil, any other failure returns `ErrCannotListSessions`
26. [TEST] #MEDIUM command-level tests with injected stdin/stdout/stderr, and a `Run` success path for each command (none exist yet; most need the runner)
27. [TEST] #EASY builders (`buildNewWindowCmd`, `buildSplitPaneCmd`, `buildRenamePaneCmd`, `buildSetLayoutCmd`): session `a#b` keeps a single `#` in `-t`, title `#S` becomes `##S` after `-T`; session names and the tmux output format are not covered yet
28. [TEST] #MEDIUM `buildNewWindowsCmd` argv table: `";"` between commands and none at the ends, `-c` omitted for an empty path, a window name with `#`, a missing panes key and an empty panes slice both return `ErrWindowWithoutPanes`; `buildSetLayoutCmd` passes the raw layout text unchanged; a window with no layout gets no `select-layout`; `appendCmd`
29. [TEST] #EASY `parseTmuxLines`: one window with one pane; two windows with one pane each (regression: every pane once landed in window 0); one window with two panes keeps pane order; two sessions that both have window 1 stay apart; `Layout` is kept only for split windows (`isSplitLayout`)
30. [TEST] #MEDIUM `parseLines` and `parseSessionLines`: names and paths with `"`, `\`, `#` and spaces come back unchanged; a row with too few fields returns `ErrCannotParseLine`; a non-numeric window or pane index returns an error that satisfies `errors.Is`; the last field keeps a stray `\x1f` (`SplitN`); empty output; `HasSession` for exit 0, exit 1 and any other failure (needs the runner seam); `Snapshot("nosuch")` returns `ErrNoSession`
31. [TEST] #MEDIUM template flow: `AsTemplate` clears the session name and every pane path and keeps windows, indexes and layouts; `Template()` returns `ErrInvalidTemplate` for 0 and for 2 sessions; `RestoreFromTempl` validates before it touches tmux (needs the runner seam); `show-template` on a file with no session returns the wrapped error; `withPaneDirPath` sets every pane path in the returned session and leaves the original unchanged (the windows and panes are cloned)
32. [TEST] #MEDIUM template names through the commands: `new-session -templ "my proj"` finds `my-proj.json` (`HasTemplate` and `ShowTemplate` agree), `save-template` and `delete-template` use the same file, a name that slugs to `""` is rejected
33. [TEST] #MEDIUM missing config dir (`t.TempDir()` plus `t.Setenv("XDG_CONFIG_HOME", ...)`): `config.New` and the read-only calls (`ListTemplates` returns an empty list, `HasTemplate` returns `(false, nil)`, `ShowTemplate` and `DeleteTemplate` return an error) create no directory; `SaveSessions` and `SaveTemplate` create it
34. [TEST] #MEDIUM config store round trip, missing file, bad JSON, traversal name (rejected by `os.Root`), non-`.json` file, delete missing, using `t.TempDir()`; `Config.SaveSessions` two consecutive calls (dir missing, then dir exists), read the file back; atomic write (`saveToFile`): first save creates no `.bak`, second save leaves `.bak` with the old content and different inodes, no `.tmp` is left behind, a failing step (for example a read-only dir) keeps the old file intact and wraps `ErrCannotSaveFile`

## Verify

Checks that need a real tmux server or a real GitHub run, not more code. Items 1 to 3 are needed before 0.8.0.

1. [VERIFY] #EASY `new-session -templ -path` (decided: `-path` is the start directory of every pane, applied by `withPaneDirPath` in `RestoreFromTempl`; without `-path` panes start in the directory `tmxu` runs from, because a `new-window` without `-c` uses the client's directory, not the session's; written, not run yet)
   - `tmxu new -templ <t> -path /tmp x`, then `tmux list-panes -s -t =x -F '#{pane_current_path}'`: every pane under `/tmp` (macOS prints `/private/tmp`)
   - `tmxu new -templ <t> y` without `-path`: panes start in the current directory
2. [VERIFY] #EASY `select-layout` in the argv chain on a multi-pane template (`ran_liftup`, `dt_demo_now` window 2): the layout is raw tmux JSON that still carries the old pane ids (`"I":"%5"`) and `"a":true`; check that tmux accepts it for the new panes, and if not strip `I` in `buildSetLayoutCmd`; tmux needs the pane count to match
3. [VERIFY] #EASY first run of `release.yaml` (written, never run on GitHub: `check` then a 4-target `build` matrix then `release`; archives `tmxu_<version without v>_<os>_<arch>.tar.gz`, `checksums.txt`, `gh release create --generate-notes`; every `v*` tag makes a normal release, there is no prerelease flag)
   - push a throwaway tag (`v0.0.1-rc1`): four green builds and a release with four archives plus `checksums.txt`
   - download one archive: `sha256sum -c checksums.txt` passes, `tar -tzvf` shows `-rwxr-xr-x` on `tmxu`, `tmxu version` prints the tag
   - `go install github.com/rogemus/tmxu@<tag>` and `tmxu version` print the tag (the `debug.ReadBuildInfo` path)
   - clean up: `gh release delete <tag> --cleanup-tag`
   - run `actionlint` on both workflows once (not installed here)
4. [VERIFY] #EASY whether `-c` (start directory) is format-expanded; escape `#` in paths if it is
5. [VERIFY] #EASY a window name or pane title ending in `;`: tmux treats a trailing `;` on an argv word as a command separator, so it may need `\;` even in argv form
6. [VERIFY] #EASY the `\x1f` field separator (`lineSafeSeparator`): check on a throwaway server that tmux prints it raw (`list-sessions -F $'#{session_name}\x1f#{session_id}' | od -c`) and what it does to a `\n` or control character in a session or window name; `parseLines` assumes neither reaches the output
7. [VERIFY] #MEDIUM on a throwaway server (private socket, `-L`): single-string chain parsing, `'it'\''s'` quoting, `new-window -k` on a session's only window, `-T` format expansion, order of `list-panes -a`, session names with `.` or `:`
   - Verified (default server, not a private socket): `new-session -s` is format-expanded (`-s 'a##b'` creates `a#b`); `-t` is not (`has-session -t '=a#b'` finds it, `'=a##b'` does not)
   - Verified: duplicate `new-session` exits 1 with `duplicate session: NAME` and stops a chain; `new-session -A -d` on an existing session fails without a terminal; `-c` with a missing directory exits 0 and falls back to `$HOME`
   - Verified: tmux 3.8 prints `#{window_layout}` as JSON; it can contain keys the structs do not model (`"l":0`), so the raw text is what gets sent back to `select-layout`

8. [VERIFY] #EASY stderr and exit code of `tmux list-sessions` when no server ever started (socket file missing): it may print `error connecting to ... (No such file or directory)` instead of `no server running`, and `ListSessions` only matches `no server`; check (`tmux kill-server; tmux ls`), then widen the match if needed

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
10. [DOCS] #MEDIUM man page `man/tmxu.1` (NAME, SYNOPSIS, COMMANDS, OPTIONS, FILES, ENVIRONMENT, EXIT STATUS, EXAMPLES, SEE ALSO)
11. [RELEASE] #MEDIUM Homebrew tap or formula, installing the man page and completions (or remove references)
12. [RELEASE] #EASY CI: `main-check.yaml` already runs build, vet, gofmt and `go test -race ./...`; add staticcheck and a `go mod tidy` diff
13. [RELEASE] #EASY add Dependabot for modules and actions (it keeps the pinned action SHAs in both workflows current)
14. [DOCS] #EASY `CONTRIBUTING.md`, `CHANGELOG.md`, `.editorconfig`
15. [TEST] #MEDIUM golden files in `testdata/` with an `-update` flag (saved JSON, help output, `ls` table)
16. [TEST] #MEDIUM optional integration test on a private socket (`tmux -L tmxu-test-<name>`), skipped without tmux
17. [TEST] #EASY tests that flags reach their fields (`-json`, `-path`, `-menu`): a value receiver or a nil `Flags()` broke this silently before
18. [TEST] #MEDIUM `SnapshotData.String` tree: last-child prefixes (`├─`, `└─`, rails), name column alignment (including multibyte names), empty pane path, several sessions separated by a blank line, windows ordered by index
19. [TEST] #EASY `printJSON`, `list-templates -json` (empty list prints `[]`, not `null`), `version -json`, `paths` and `paths -json` (three lines in the order config, templates, sessions; creates nothing)
20. [TEST] #EASY `Info()` of the six commands that format `DescLong` from `cfg` (`%s` is filled with the real path, no stray `%!s(MISSING)`)
21. [TEST] #MEDIUM `utils.Confirm` (build the input with `bufio.NewReader(strings.NewReader(...))`; cover `y`, `n`, EOF with data, EOF with no data) and `utils.RequireTerminal` (non-file readers pass), `Line` conversions, pane builders, `buildNewSessionCmd`, `newSessionError`
22. [TEST] #MEDIUM window layout: decoding real tmux text, `isSplitLayout` (single pane, split, empty, invalid), `Data.String` and `Data.JSON` output, `buildSetLayoutCmd`, `select-layout` after the splits
23. [TEST] #EASY `config.New` (`t.Setenv` for `XDG_CONFIG_HOME` and `HOME`, version trimming, missing home)
24. [TEST] #EASY `internal/cli` help output (order, aliases), command lookup, `ErrUnknownCommand`, `-h` handling
25. [TEST] #MEDIUM `Cli.Run` adds `\nUsage: ...` to an `ErrUsage` error and keeps `errors.Is(err, cmd.ErrUsage)` true; `Info.Usage()` with and without `Arg`; `new-session` on a duplicate session returns an error that satisfies `errors.Is(err, tmux.ErrSessionExists)` (needs the runner seam)
26. [TEST] #EASY use `t.Helper`, `t.Cleanup`, `t.Setenv` in tests (`just test` already runs with `-race`)
27. [FEATURE] #HARD performance (parked): symptom is panes "catching up" (shells still booting, prompts redraw) after restore; baseline first with `time tmxu restore`, and if it is under about 2 s the first two items are probably enough
    - `new-session -x <cols> -y <rows>` from `term.GetSize`; no flags when stdout is not a TTY
    - Wait for panes to be ready: poll `capture-pane -p -t <pane>` every ~50 ms, ~5 s timeout per session, sessions in parallel; `waiting for shells...` on stderr; warn on timeout; `-no-wait` to skip
    - Batch a window's tmux calls into one invocation (low priority)
    - Staggered start in batches of 3-4 sessions (only if still janky)
    - Parallel snapshot with `sync.WaitGroup`, indexed result slice and a semaphore (measure first); restore stays sequential

## 0.9.0 plans

Planned for the release after 0.8.0. The man page and the Homebrew tap in Nice to have depend on completions.

1. [FEATURE] #MEDIUM `tmxu completion zsh|bash|fish` (hand-written, `//go:embed` or generated from the registry); hidden `__complete` for session and template names; tests for it; a README section on installing completions
2. [DOCS] #EASY README fzf recipe: `tmxu ls -json | jq -r '.sessions[].session_name' | fzf`, then `tmxu attach`; a shell function and a tmux `bind-key` example; `jq` is optional; `attach -menu` is the built-in alternative
3. [BUG] #HARD no rollback (deferred from 0.8.0 on purpose): a failed window step leaves a half-built session that blocks the retry with `duplicate session`; applies to `new-session -templ` and to restore (roll back sessions created in this run when a later step fails; `tmux.RestoreSession` is the one place both go through, so the rollback belongs there)
4. [FEATURE] #MEDIUM read the real `base-index` and `pane-base-index` (restore targets `=S:W.<saved pane index>`, so it only works when the restoring server uses the same bases as the saving one; `BaseIndex` was deleted as unused; `tmux show-options -gv base-index` reads the global option, the session's own may differ); handle saved index != current base so no blank leftover window stays

## Cosmetics

Naming, text, structure, duplication, doc comments. No behavior change for users.

1. [COSMETIC] #EASY `helpCmd`: align the FLAGS descriptions in a column (`-path` and `-templ` differ in length, the template pads with fixed spaces); use `text/tabwriter` like `listAllCommands`, with a tab between name and description, flushed after `Execute`; check that the other sections are not affected
2. [COSMETIC] #MEDIUM doc comments on exported identifiers (package docs exist everywhere; `Info`, `Flag`, `IOs`, `Cmd`, `Config`, `New`, `SnapshotData`, `SessionData`, `WindowData`, `PaneData`, constructors and `utils.Slug` have none); unexport what is not API
3. [COSMETIC] #EASY `HasTemplate` duplicates what `ShowTemplate` reports; keep it, or return a not-found sentinel from `ShowTemplate`
4. [COSMETIC] #EASY config: fold the repeated `os.OpenRoot` + `defer root.Close()` (5 places: four in `templates.go`, one in `sessions.go`) into one helper
5. [COSMETIC] #MEDIUM generic `readJSON[T]` / `writeJSON`; `os.MkdirAll` (file reads, writes and deletes already go through `os.Root`: `readFile`, `saveToFile`, `deleteFile` in `tmxu_config.go`)
6. [COSMETIC] #EASY `Root`-based listing for `ListTemplates` (`fs.ReadDir(root.FS(), ".")`; `os.Root` has no `ReadDir`; `readTemplatesName` then takes `[]fs.DirEntry`)
7. [COSMETIC] #EASY a window with more than about 7 panes fails restore with `no space for a new pane` (each `split-window -h` halves the active pane, whatever the window size; accepted as user error). If it ever matters: `select-layout even-horizontal` after each split, then the saved layout (checked on tmux 3.8: an 80x24 detached window reaches 21 panes and the saved layout applies as is)

## Settled decisions

Window layouts are stored as raw tmux JSON and applied with `select-layout` after the splits (the normalized `LayoutNode` code is deleted). Errors, prompts and diagnostics go to stderr, data and success messages to stdout. A window without panes is an error (`ErrWindowWithoutPanes`, no `panic`). One argv chain with `";"` elements (no `source-file`). Derived pane names `<window>_<index>`. On-disk format is the nested `SnapshotData`: `{"version":1,"sessions":[{"session_name","session_windows":[{"window_name","window_index","window_layout"?,"window_panes":[{"pane_index","pane_path"}]}]}]}`; no slugs, no tmux ids, no version check and no migration (never released). A template is a snapshot with exactly one session whose name and pane paths are cleared (`AsTemplate`); `SnapshotData.Template` validates that and `RestoreFromTempl` builds the session from it. `dump -session` prints real data (sanitizing lives in `save-template`). Text output is a names-only tree (no ids, indexes or layouts; those stay in `-json`). JSON shapes are `{"sessions":[...]}`, `{"templates":[...]}`, `{"version":"..."}` and `{"config","templates","sessions"}` (paths). Flags have one name each (`-yes`, no short alias; `flag` has none). Restore with sessions that already exist: refuse by default; `-force` kills only the saved sessions that already exist, then recreates them.
