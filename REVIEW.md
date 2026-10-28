# Review

Review of `tmxu` (about 1.6k lines of Go). Short checklist: [TODO.md](TODO.md).

Policy: **stdlib-first**. The goal is to learn Go by writing the code by hand, so each fix below uses the standard library. Every section ends with an optional "library alternative" for when you want to trade learning for speed. `golang.org/x/term` (already used for raw mode) is the only dependency kept.

Baseline: `go vet`, `gofmt -l` and `go test ./...` are clean (16 tests). Findings come from manual reading; staticcheck could not run (Go 1.27 export data error), so run it yourself. Line numbers refer to the code at the time of review and will drift.

---

## 1. Bugs

These are real defects, not style. Fix them first and add a test for each.

### 1.1 `TimeSince` reports wrong units (`cli/time.go`)

- `case seconds < oneMin` is unreachable because `seconds < 60` is caught earlier. Minutes are never reported.
- The unit labels are shifted: a branch under one hour divides by `oneHour`, giving 0, and prints "1h ago". The menu shows "1h ago" for a 5-minute-old session.
- `oneMonth` is 28 days and `oneYear` is 336 days; they are package `var`s with redundant `int()` conversions.
- **Fix:** keep a `time.Duration` and `switch` on `d < time.Minute`, `d < time.Hour`, `d < 24*time.Hour`, and so on. Write a table test first.
- **Learn:** `time.Duration` arithmetic, table-driven tests.

### 1.2 `newSemVer` ignores the real major version (`cli/semVer.go:16-19`)

- `strings.Split(v, ".")` runs before `strings.TrimPrefix(v, "v")`, so the trimmed value is dead and `"v0.7.0"` parses major as `Atoi("v0") = 0`. `v1.x` and `v2.x` both compare as major 0.
- `Atoi` errors are discarded (`semVer.go:22-28`), so a wrong parse is silent.
- **Fix:** trim first, then split, check `len(parts) == 3`, return errors, and add a `Compare(other) int` method. Compare versions instead of strings in `cmd.go:152-153`, otherwise a locally newer dev build is told to "update".
- **Library alternative:** `golang.org/x/mod/semver`.

### 1.3 `versions.go` is fragile

- `io.ReadAll` error ignored; `fmt.Printf("[%v]\n", res)` prints a debug value; `http.Get` has no timeout and no status check.
- `getNewestVersion` panics on an empty tag list (`sorted[len(sorted)-1]`; `strings.Split("", "\n")` yields `[""]`).
- The `version` command hits the network on every call, which makes `tmxu --version` unusable offline and in scripts. The comment says GitHub, but the code queries `proxy.golang.org`.
- **Fix:** `http.Client{Timeout: 3 * time.Second}` with `http.NewRequestWithContext`, check `StatusCode`, wrap the body in `io.LimitReader`, handle empty results. Make the check opt-in (`version --check`) and degrade to the local version on failure.

### 1.4 `HasSession` treats "no" as an error (`cli/tmux.go:150,199-211`)

- `tmux has-session` exits 1 when the session does not exist, so the normal "no" case returns an error. `NewSession` discards it (`hs, _ :=`) and only works by accident. `cmd.go:367-369` cannot tell "absent" from "tmux failed". `hs == true` is non-idiomatic.
- **Fix:** run through one helper that returns a typed error, then `errors.As(err, &exitErr)` and treat exit code 1 as `(false, nil)`.

### 1.5 Panics from unchecked indexes

- `tmux.go:54-58` `newTSession` reads `parts[2]` without a length check. Empty tmux output gives `[""]`, so `cmd.go:69` panics too.
- `tmux.go:163` and `:238` read `Windows[0].Panes[0]`, which panics on hand-edited JSON with a window without panes.
- `menu.go:43`: `strings.Repeat("·", 25-len(title))` panics for a title longer than 25 bytes, and counts bytes, not runes.
- `menu.go`: Enter on empty `items` indexes `items[selected]`.
- **Fix:** validate lengths and return errors. Parse each tmux line once into a typed struct (see 5.2).

### 1.6 Menu is unsafe outside a TTY (`cli/menu.go`)

- `term.MakeRaw` error is ignored (line 21), so a non-TTY passes a nil state to `Restore`.
- `ReadByte` error is ignored (line 50): on EOF this is a 100% CPU busy loop. The `reader.Read` result at line 55 is also ignored, and the ESC detection through `Buffered() > 0` is racy.
- In raw mode Ctrl-C is byte `0x03`, not SIGINT, and the switch has no case for it. Only `q` and Esc quit.
- The menu clears the screen with `\033[H\033[2J`, destroying the user's scrollback; the escapes are repeated string literals.
- **Fix:** check `term.IsTerminal` on stdin and stdout and fall back to a plain list. Handle EOF, `case 3, 4`. Use constants for escapes and the alternate screen (`\033[?1049h`/`l`). Draw the menu on stderr or `/dev/tty` so stdout stays clean for the selection.

### 1.7 Restore and base indexes

- `cmd.go:136,189`, `tmux.go:278` assume `base-index` is 1; `NewPane` hard-codes `pane.Order != 1`; `cmd.go:522` uses `i+1`. With a different `base-index` or `pane-base-index`, panes are misaddressed.
- `RenamePane` runs `select-pane`, which changes the active pane as a side effect during restore.
- Template layouts are saved but `new-session -templ` does not call `SetWindowLayout`, while `restore-sessions` does. Decide whether that is a bug.
- **Fix:** read `base-index` and `pane-base-index` via `show-options -gv` once, or address by the index tmux returned. Set pane titles without `select-pane` (`select-pane -T` with `-t`).

### 1.8 Input validation

- `new-session` with an empty name is not validated (`cmd.go:488-512`), unlike `attach` and `save-template`.
- Template names are used unsanitized in paths (`config.go:146,175,211,227`), so `save-template -name ../x` writes outside the templates directory.
- `loadTemplateFiles` reads every directory entry (including non-`.json` and subdirectories) and fails the whole list on one bad file.
- `delete-template` reads `os.Args[2]` directly, with no flag parsing.
- **Fix:** a `validName(string) error` (reject empty, `/`, `\`, `..`, leading `.`). Skip non-`.json` entries and subdirectories, and report bad files without failing the list.

### 1.9 Destructive behavior

- `restore -force` kills existing sessions, including the one the user is in (`tmux.go:154-159`), which can kill their terminal. The prompt is generic and does not list what will be killed.
- `confirm` (`cli.go:111-117`) creates a new `bufio.Reader` per call, drops the read error, and returns false on EOF, so cron or piped use silently prints "Aborted." (and has inconsistent messages "Aborted." vs "Aborted \n").
- A half-created session after a mid-way failure stays behind and blocks the next run.
- **Fix:** detect the current session (`$TMUX`, `display-message -p '#S'`) and refuse; list sessions to be killed in the prompt; add `-y/--yes`; if stdin is not a TTY and no `--yes`, return an explicit error; roll back the partial session on failure.

---

## 2. Error handling

### 2.1 No wrapping, capitalized, trailing `\n`, causes dropped

- No `%w` anywhere. Typical: `fmt.Errorf("Unable to read ... \n")`, `fmt.Errorf("Unable to create tPane: %s \n", tp.Name)`, where `tp.Name` is empty because the call failed. Files: `config.go` (about 30 sites), `cmd.go`, `tmux.go`, `versions.go`. Staticcheck ST1005 flags these.
- Root causes are lost: a failed `os.ReadFile` reports only "Unable to read" with no reason (permission denied vs not found). Mid-restore failures only show the outer message.
- `cli.Run` (`cli.go:101`) prints with `%s` and no newline, which is why every message carries a hand-added `\n`.
- **Fix:** `fmt.Errorf("read sessions file %s: %w", path, err)`. Lowercase, no trailing punctuation or newline. Add the newline in exactly one place: `fmt.Fprintln(os.Stderr, "tmxu:", err)`. Return or wrap `err` instead of replacing it with a generic message (`cmd.go:63,186,197,226,264,275,516`).
- Typos in messages: "sassion", "abborded", "availabe", "Widnows", and `"%s ]n"` at `config.go:67`.
- **Learn:** `%w`, `errors.Unwrap`, why wrapping beats string concatenation.

### 2.2 Swallowed errors

`tmux.go:29,32,141,142,150`, `semVer.go:22-28`, `versions.go:19`, `menu.go:21,50,55`, `cli.go:114`, `config.go:109,123,210,226` (`path, _ := getTemplatesDirPath()`). A wrong-but-silent parse is worse than a failure. Check every one.

### 2.3 Sentinels and typed errors

- Existing sentinels `errorSessionExists` and `errorAborded` break the Go naming convention. Rename to `errSessionExists`, `errAborted` (or exported `Err...` if needed outside the package).
- Add `ErrNoSession`, `ErrNoTemplate`, `ErrNoTmux`. Detect a missing binary with `errors.Is(err, exec.ErrNotFound)` and print "tmux not found in PATH". Today a missing tmux reads "unable to list tmux sessions".
- Add a typed error from one `run()` helper:
  ```go
  type TmuxError struct {
      Args   []string
      Stderr string
      Err    error
  }
  func (e *TmuxError) Error() string { ... }
  func (e *TmuxError) Unwrap() error { return e.Err }
  ```
  Callers use `errors.As` for `*exec.ExitError` and check `ExitCode()`. tmux's own stderr then shows up in messages.
- Replace `os.Stat` + `os.IsNotExist` (`config.go:52,110,228`) with `errors.Is(err, fs.ErrNotExist)` and call `os.Remove` directly: check-then-act is a TOCTOU pattern.
- Use `errors.Join` to collect multiple restore errors.

### 2.4 Exit, stderr and stdout

- `cli.Run` calls `os.Exit` in five places (`cli.go:85,90,95,102,107`), so a method named `Run` is not testable. There is no `log.Fatal` or deliberate `panic`; the panics are accidental (1.5).
- Errors go to stdout via `fmt.Printf`. `version` prints partial output and then returns an error (`cmd.go:148-149`).
- Unknown command prints "Invalid command " (trailing space, no name). `-h`/`--help` fall through to "Invalid command". `help` is handled with a `len(os.Args)==N` chain (`cli.go:88-96`).
- **Fix:** `func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error`, with a single `os.Exit` in `main.go` that maps errors to codes via `errors.As` on a typed `usageError`. Codes: 0 success, 1 runtime failure, 2 usage error, 130 interrupt. Print only data to stdout; errors, prompts, "Aborted" go to stderr. Handle `flag.ErrHelp` (exit 0, help to stdout).
- Inject `io.Writer` instead of calling `fmt.Print*` directly (`cmd.go:66-69,148-161,229,272,297,320-330,420,445,550`, `menu.go`, `output.go:24-27`); that makes output testable and golden-file friendly.

---

## 3. Architecture and structure

### 3.1 Global state and `os.Args`

- Commands read `os.Args[2:]` (`cmd.go:93,253,358,482,436,440`), so they cannot be tested without mutating the global. `version` is a package-level var set as a side effect in `NewCli` (`cli.go:10`). `Cmd` values are package-level vars used only by `NewCli`. `NewCli` returns an unexported `*cli` from an exported function.
- **Fix:** `Run func(args []string) error` on each command; keep `version` on the struct; build commands in a constructor or as methods. Export the type or return an interface.

### 3.2 Package layout and naming

- Everything is in one package `cli`, with exported functions that are not an API (`ListSessions`, `NewSession`, `AttachToSession`, `Max`, `Min`, `TimeSince`). Exported functions take unexported types (`NewSession(session tSession...)`).
- `t` prefixes (`tSession`, `tWindow`, `tPane`, `tTemplate = tSession`) are non-idiomatic. Rename to `Session`, `Window`, `Pane`, `Template`. `tSessionSimple` (and its `Widnows` typo) is a view model that belongs next to the menu. `newTSession` returns an `error` that is always nil and ignores `parts[0..1]`.
- `semVer.go` -> `semver.go`; merge `versions.go` into `update.go`.
- `math.go`: delete; Go 1.21+ has builtin `min` and `max` (`menu.go:60,62`).
- `output.go`: `renderTable` pads by `len(cell)` (bytes) and breaks on multibyte names; use `text/tabwriter` (5 lines).
- `main.go`: `cli := cli.NewCli(...)` shadows the package name.
- `int16` for `Order` (`tmux.go:15,63,89`) with truncating `int16(order)` casts after `strconv.Atoi`; use `int`.
- `cmd.go` is 553 lines with `Run` closures of 70-80 lines (restore, new-session). Split per command, and extract helpers (section 4).
- **Decision:** keep the flat `cli/` (fine at 1.6k lines) or split into `internal/tmux`, `internal/store`, `internal/cli`. Keep `main.go` at the repo root so `go install github.com/rogemus/tmxu@latest` keeps working.

### 3.3 Documentation

- Almost no godoc. Exported identifiers (`Cmd`, `NewCli`, `ListSessions`, `NewSession`, `AttachToSession`, `HasSession`, `NewWindow`, `RenameWindow`, `NewPane`, `RenamePane`, `TimeSince`) are undocumented, and there is no package comment. Unexport what is not API, document the rest.

---

## 4. Duplication

| What                                         | Where                                                                                                                                                                                                            | Fix                                                                                                                                                                          |
| -------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Session/window/pane scan loop                | `cmd.go:189-222` (save-sessions) and `377-408` (save-template)                                                                                                                                                   | One `snapshotSession(ctx, name, order) (Session, error)`; template path calls `toTemplate(s)`                                                                                |
| Restore loop (window, then panes)            | `cmd.go:278-295` and `538-548`                                                                                                                                                                                   | One `restoreSession(ctx, s, force)`                                                                                                                                          |
| `errors.Is(err, errorSessionExists)` mapping | `cmd.go:271-276, 505-509, 532-536`                                                                                                                                                                               | One helper                                                                                                                                                                   |
| `"TEMP_VALUE"`                               | local const at `cmd.go:372`, refill at `520-529` not using it                                                                                                                                                    | Package const, or remove the placeholders by dropping the fields from `Template`                                                                                             |
| Config I/O                                   | `config.go` marshal/write/read/unmarshal repeated for sessions and templates; `os.UserHomeDir()` about 9 times; `hasX` + `createX` pairs                                                                         | `readJSON[T any](path) (T, error)`, `writeJSON(path, v) error`, one `configDir()`, `os.MkdirAll` instead of has/create                                                       |
| Path building                                | `fmt.Sprintf("%s/%s.json")` (`config.go:146,175,211,227`); `configDir = "/.config/tmxu/"` with odd slashes                                                                                                       | `filepath.Join`                                                                                                                                                              |
| Literal `~/.config/tmxu` in messages         | `cmd.go:170,226,229,238,308,316,342,420,429,442,445`                                                                                                                                                             | Print the resolved path                                                                                                                                                      |
| Flag/help text                               | flags declared in `Cmd.Flags [][]string` and again in `fs.BoolVar`/`StringVar` with different wording (`cmd.go:239/251, 344/356, 456/479`); README is a third copy; `Cmd.Examples` stored but never printed      | Build the `FlagSet` once per command and generate help from `fs.VisitAll`; define a `Flag` struct instead of `[][]string`; print examples or delete them                     |
| `SessionName`/`SessionWindow`                | copied into every `Window` and `Pane` (`tmux.go:62-94`) although derivable (`name:index`)                                                                                                                        | Remove from persisted structs; compute targets at restore time (`fmt.Sprintf("%s:%d", ...)`). This changes the JSON schema, so add a `"version"` field and migrate old files |
| Raw tmux line parsing                        | `strings.SplitN(s, "\t", 3)` in `cmd.go:68`, `tmux.go:27,54,72,107,139-140`, each with different bounds checks; `ListSessions() []string` returns tab-joined strings and the sort re-parses them each comparison | Parse once into typed structs with `Created time.Time`                                                                                                                       |
| Splitting output                             | `strings.Split(strings.TrimSpace(string(output)), "\n")` (`tmux.go:134,194,272`, `versions.go:21`)                                                                                                               | One helper (`lines(out []byte) []string` that returns nil for empty)                                                                                                         |
| `exec.Command("tmux", ...)`                  | 12 sites (`tmux.go:129,155,167,176,189,200,239,249,258,267,281,297`)                                                                                                                                             | One `run()` helper (and the test seam, 5.1)                                                                                                                                  |

Typos and wrong text found along the way: the `list` output says "with their IDs and names" but prints an index and a name; examples use "tmux" where "tmxu" is meant (`cmd.go:84,174,244-246`); `new-session` flag set is named `"new-session"` inside save-template (`cmd.go:355`, copy-paste), so flag errors name the wrong command; `tTemplate(ts)` at `cmd.go:415` is a no-op conversion on an alias.

Small style nits worth fixing while you are there:

- `if hs == true && !force {...} else if hs == true && force {...}` (`tmux.go:152-154`) -> `if hs { if !force {...}; kill }`.
- `else` after `return` (`config.go:52-56,110-114`, `tmux.go:228-245`); `if err := cmd.Run(); err != nil { return err }; return nil` -> `return cmd.Run()` (`tmux.go:181-185,287-292`).
- `args := append(startDirArgs(...), "-t", ...)` (`tmux.go:238`) can alias the helper's backing array. Safe today, fragile.
- `numberOfPane` (`cmd.go:268`) is misnamed and declared at the wrong scope.
- Flag sets use `ContinueOnError` but discard the parse error, so `-badflag` prints stdlib usage to stderr plus "unable to parse flags". Return the error with `%w` and set `fs.SetOutput`.
- Alias collisions are not detected (`s`, `st`, `dt`, `lt`, `v`), and `cmdsOrder` duplicates `cmds`.
- `Cmd` is exported but package-internal; its `Arg` field is a single string.

### 4.1 The saved sessions file is redundant (measured)

Measured on a real `~/.config/tmxu/tmux-sessions.json`: 15 sessions, 58 windows, 59 panes, 33,291 bytes.

| Field                           | Redundant in                        | Derivable from                         |
| ------------------------------- | ----------------------------------- | -------------------------------------- |
| `window.sessionName`            | 58/58                               | `session.name`                         |
| `window.sessionWindow`          | 58/58                               | `session.name:window.order`            |
| `pane.sessionName`              | 59/59                               | `session.name`                         |
| `pane.sessionWindow`            | 59/59                               | parent window                          |
| `order` (session, window, pane) | 132/132 equal to array index + 1    | array index                            |
| `pane.name`                     | 57/59 equal to `window.name`        | `window.name` (`RenamePane` copies it) |
| `layout` on single-pane windows | 57/58 windows have exactly one pane | not needed (see below)                 |

Dropping the four `session*` fields, `order` and the redundant `pane.name` shrinks the file by about 46% (33,291 to 17,841 bytes) without changing restore behavior. Storing the same fact in several places means the file is only valid while every copy agrees; it is also why templates need the `TEMP_VALUE` hack. Removing the copies makes inconsistent state impossible to write.

Keep `order` only if you want to preserve tmux's real window indexes (gaps from closed windows); otherwise use the array index and read the real `base-index` (see 1.7). Keep `pane.path` explicit per pane: 16 distinct values over 59 panes looks repetitive, but inferring it from the window is fragile.

Target shape:

```json
{
  "version": 2,
  "sessions": [
    {
      "name": "ai_devrel_knowledgebase",
      "windows": [
        {
          "name": "editor",
          "panes": [{ "path": "/Users/me/Developer/project" }]
        },
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

Wrapping the array in an object with a `version` field lets you tell old and new files apart. `encoding/json` ignores unknown keys, so old files still load; only the four dropped fields need a one-time reader if you want to verify them.

#### Layout: new tmux only, real JSON, multi-pane only

- **Double encoding.** `layout` is stored as a string containing escaped JSON (`"{\"V\":2,\"L\":{...}}"`). tmux 3.8 prints its layout as JSON (`#{window_layout}`), and `tmxu` stores that output verbatim in a Go `string`, so `encoding/json` escapes it. All 58 layouts in the measured file parse as valid JSON.
- **Decision: support new tmux only (3.8+).** Older tmux emits the checksum form (`d2b3,198x53,0,0,1`), which is not JSON. Drop support for it. That removes the need for a dual-form marshaller.
- **Store it as `json.RawMessage`**, so the file contains a real nested object instead of an escaped string. Validate with `json.Valid` when saving and fail with a clear error if tmux returns something else. The layout stays opaque: you never interpret it, only pass it back to `select-layout` on restore.
- **Omit it for single-pane windows.** A single pane fills the window, and the window size comes from the terminal you restore into, not from the saved layout. Tag the field `json:"layout,omitempty"`; a `json.RawMessage` that is nil or empty is omitted. In the measured file this removes the field from 57 of 58 windows.
- **Do not call `select-layout` on restore** when a window has no layout.
- **Why `omitempty` and not a pointer:** `json.RawMessage` is a `[]byte`, so the zero value is omitted cleanly.
- **Pane ids are runtime-only.** The layout contains `"I":"%0"` pane ids that mean nothing after a restart. Treat the layout as opaque and let tmux remap panes when `select-layout` is applied.
- **Check the version at startup** with `tmux -V` and document the minimum in the README.

---

## 5. Testing

Current state: `cli/tmux_test.go` is the only test file. Two table-driven tests with subtests, well written, but they only cover `paneName` and `newTPane`. No `t.Helper`, `t.TempDir`, `t.Setenv`, fakes or golden files. CI does not run tests at all, so `-race` is never used.

### 5.1 The seam: a `Runner` interface

Everything shells out to tmux directly, so almost nothing is testable. Introduce:

```go
type Runner interface {
    Run(ctx context.Context, args ...string) ([]byte, error) // capture stdout
    Attach(ctx context.Context, args ...string) error        // inherits stdio
}

type execRunner struct{}

type Tmux struct{ r Runner }
```

Tests inject a `fakeRunner` with a `calls [][]string` field and a `fn func(args []string) ([]byte, error)`, then assert exact args, e.g. `[]string{"new-window","-c",dir,"-t","s:2","-n","w"}`. The same seam gives you `-v` (log args in one place) and `--dry-run` (print instead of run) almost for free.
**Learn:** small interfaces defined where they are consumed, dependency injection without a framework.

### 5.2 Typed parsing

Parse tmux output once into `SessionInfo{Name string; Created time.Time; ...}`, `Window`, `Pane`. Remove the tab-joined string APIs. Parsers return errors on short lines.

### 5.3 Tests worth writing

- **Pure unit tests (no seam needed):** `TimeSince`, semver parse/compare/sort, `renderTable`.
- **Tmux layer with the fake:** `ListSessions` ordering by `session_created`; empty output (currently panics); `NewSession` with and without force; `startDirArgs` (use `t.TempDir()` for an existing dir and a regular file for the "not a directory" case, `tmux.go:215`); `NewWindow` first vs later windows.
- **Config:** make a `store{dir string}` so no `$HOME` mutation is needed (otherwise `t.Setenv("HOME", t.TempDir())`). Cases: round trip, missing file, malformed JSON, template name with a path separator (must be rejected), a non-`.json` file in the templates dir, delete of a missing template. Compare with `reflect.DeepEqual`.
- **Golden files** in `testdata/` with an `-update` flag: the saved sessions JSON (pins the on-disk schema before you drop `SessionName`/`SessionWindow`), `help`/`helpLong` output (needs writer injection), `renderTable`.
- **Template placeholder logic** (save substitutes, new-session refills) once the snapshot code is out of the closure.
- **Command level:** with `Run(args)`, the fake runner, and an injected `io.Reader` for `confirm`, run `save`, `restore`, `new -templ` end to end.
- **Optional integration test:** `exec.LookPath("tmux")` or `t.Skip()`, with a private socket `tmux -L tmxu-test-<name>` so it never touches real sessions. The runner needs to accept global prefix args.
- Helpers call `t.Helper()`; temp state uses `t.Cleanup`.
- Run `go test -race ./...` in CI. It is cheap and becomes essential once goroutines appear.

**Library alternative:** `github.com/google/go-cmp/cmp` for diffs, `testify` for assertions (not needed).

---

## 6. Context, signals and concurrency

### 6.1 Context and timeouts

- No `context` anywhere. Add `ctx context.Context` as the first parameter of everything that shells out or does network I/O. In `main`: `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM); defer stop()`.
- `exec.Command` -> `exec.CommandContext(ctx, "tmux", ...)`. For non-interactive calls wrap with `context.WithTimeout(ctx, 5*time.Second)` so a hung tmux server socket cannot block the CLI forever. Do **not** put a timeout on `AttachToSession` (interactive, long-lived); pass only the signal context there.
- HTTP: see 1.3.
- `CommandContext` kills the child on cancel, so Ctrl-C mid-restore leaves a half-restored session. Decide: clean up, or report which session was in progress.

### 6.2 Raw mode and signals

`menu.go:22` uses `defer term.Restore(...)` and a deferred cursor show. Defers do not run on `os.Exit` or a fatal signal, so SIGTERM/SIGHUP leave the terminal raw with the cursor hidden. Return through normal control flow on context cancellation so defers run. The blocking `ReadByte` needs a goroutine (or closing the fd on cancel) to be interruptible. Handle EPIPE/SIGPIPE gracefully for `tmxu ls | head -1`. Exit 130 on interrupt.

### 6.3 What can run concurrently

Today there are no goroutines, so there are no data races.

**Safe (read-only):** listing windows per session and panes per window in `save-sessions` / `save-template` (`cmd.go:189-222, 377-408`). After extracting `snapshotSession`:

```go
out := make([]Session, len(infos))
var wg sync.WaitGroup
sem := make(chan struct{}, 4)
for i, s := range infos {
    wg.Add(1)
    sem <- struct{}{}
    go func() {
        defer wg.Done()
        defer func() { <-sem }()
        out[i], errs[i] = snap(ctx, s)
    }()
}
wg.Wait()
```

Write results by index (`out[i]`, `errs[i]`) so no mutex is needed and order is deterministic. Never `append` to a shared slice from goroutines. Go 1.22+ per-iteration loop variables make the closure capture safe (`go.mod` requires 1.26). Gain: each tmux call is a ~5-10 ms process fork, so this only matters with dozens of windows. Measure before keeping it; as a learning exercise it is excellent.

**Possible but risky: restoring independent sessions** (`cmd.go:267-298`). Risks: interleaved output; the first `new-session -d` starts the tmux server, and racing starts can fail with "server exited unexpectedly" (create the first session sequentially, fan out the rest); shared global option state; error handling (first-error-cancel vs collecting with `errors.Join`).

**Must stay sequential:**

- Window and pane creation inside one session. `new-window` assigns the next index and `split-window` targets an index that exists only after the previous call; layout depends on panes already existing (`cmd.go:278-295`).
- `NewSession`, then first-window rename, then panes (`tmux.go:149,226`).
- `RenamePane` (uses `select-pane`, changes the active pane, `tmux.go:295`).
- Sorting in `ListSessions` (after all rows are gathered; a single call).
- The interactive menu and `attach`.

Recommendation: context and timeouts (high value); the Runner seam and tests (high value); parallel snapshot (optional learning exercise); restore stays sequential for correctness.

**Library alternative:** `golang.org/x/sync/errgroup` (`SetLimit`, first-error cancellation) replaces the WaitGroup and semaphore boilerplate.

---

## 7. Unix-style CLI

Baseline: hand-rolled dispatch in `cli/cli.go:82-109` on `os.Args`; per-command `flag.NewFlagSet` with `ContinueOnError`; only `help` and `version` subcommands; no `-h`, `--help`, `-V`, `--version`; exit code 0 or 1; errors and output both on stdout; no completion, color, quiet/verbose, JSON or dry-run; config path hard-coded; files written `0644` and dirs `0755`; no man page, no Formula, no CONTRIBUTING or CHANGELOG (LICENSE exists).

### 7.1 Help, version and flag conventions

- `tmxu -h`, `tmxu ls -h`, `tmxu --help` print "Invalid command" plus full help and exit 1. Handle `flag.ErrHelp`: help to stdout, exit 0. Usage errors go to stderr, exit 2.
- `-V`/`--version` globally; the `version` subcommand is local (print `tmxu 0.7.0`, one line, no trailing space; `cmd.go:148,154,161` print a trailing space). Network check behind `version --check`, and `TMXU_NO_UPDATE_CHECK`.
- Help layout: Usage, Commands, Options, Environment, Exit status. Print `Examples`.
- stdlib `flag` accepts `-flag` and `--flag` identically but has **no short/long pairs, no combined short flags, and stops at the first positional argument**. So `tmxu new mysession -path /tmp` silently treats `-path` as a positional argument and ignores it. Fix by documenting "flags before arguments" and detecting stray `-x` tokens in positional args, or by pre-splitting `args` into flags and positionals yourself. Support the `--` terminator so a session named `-x` works (the `flag` package stops at `--`; remember to honor it in any pre-splitting).
- Missing required argument: print usage, exit 2 (today `new-session` creates an unnamed session).
- Unknown command: suggest the nearest (small Levenshtein, ~15 lines).
- `tmxu` with no arguments prints help and exits 0 today. Pick one behavior deliberately (exit 2 with usage on stderr, or an interactive picker on a TTY).
- **Library alternative:** `spf13/cobra` gives short/long flags, `--`, did-you-mean, `completion` and man generation (`cobra/doc`) and deletes `Cmd`, `help`, `listAllCommands` and `renderTable`, at the cost of ~3 MB of dependencies and a more opinionated structure. `spf13/pflag` is the lighter option (flags only).

### 7.2 Exit codes

0 success, 1 runtime failure, 2 usage error, 130 interrupt; consider a distinct code (127) for "tmux not installed". Document them in help and the man page. Declining a prompt exits 0 today; if non-interactive, exit 1 so scripts can tell. Implement with a typed `usageError` and `errors.As` in `main`.

### 7.3 stdout/stderr, quiet/verbose

Only data on stdout (session list, template list, version). Print the "Available sessions" header only on a TTY (`cmd.go:66` pollutes pipes). Add `-q` (suppress "Session created..." at `cmd.go:229,297,420,445,550`) and `-v` (log each tmux command, trivial with the Runner). Use `log/slog` writing to stderr instead of `fmt.Printf`.

### 7.4 Color and TTY

No color today. If added: only when stdout is a TTY (`term.IsTerminal(int(os.Stdout.Fd()))`), `NO_COLOR` unset, `TERM != "dumb"`, and honor `--color=auto|always|never`. The menu never checks for a TTY (1.6). `confirm` reads stdin without a TTY check (1.9): add `-y/--yes`, and error with "refusing to prompt: stdin is not a terminal; pass --yes".

### 7.5 Machine-readable output

- `ls` prints a numbered padded table (`cmd.go:66-70`). Add `--format plain|json` (or `--json`): one name per line when stdout is not a TTY, and `-0/--null` for `xargs -0`. Same for `list-templates` (`cmd.go:324-332`).
- `attach -` reads a name from stdin so `tmxu ls | fzf | tmxu attach -` works; at minimum document `tmxu attach "$(tmxu ls --plain | fzf)"`.
- `save -o -` and `restore -f FILE` for stdin/stdout and alternate paths (also makes tests easier); `show-template NAME` prints JSON.
- The menu UI should be drawn on stderr (like fzf) so a "print selection" mode can use stdout.

### 7.6 Paths and environment

- Replace `"/.config/tmxu/"` (`config.go:10`). **Do not use `os.UserConfigDir()` blindly**: on macOS it returns `~/Library/Application Support` and would silently strand existing users' data. Implement XDG explicitly (`$XDG_CONFIG_HOME`, else `~/.config`), or add a fallback read of the old path.
- Optional XDG purism: saved sessions are state (`$XDG_STATE_HOME` or `$XDG_DATA_HOME`), templates are user-authored config. Needs a migration, so decide deliberately.
- `TMXU_CONFIG_DIR` and `--config-dir` (precedence flag > env > default; useful for tests). Do not clear the environment: `TMUX_TMPDIR` and `TMUX` are consumed by tmux itself.
- `tmxu paths` prints the resolved directories (scripting, bug reports). Print real paths in messages (section 4).
- Document an "Environment" section in help and the man page.

### 7.7 tmux awareness

- `attach` (`tmux.go:175-186`) runs `tmux attach -t NAME`. Inside tmux (`$TMUX` set) that fails or nests; use `switch-client -t NAME` instead. Same for `-menu`.
- `-t foo` prefix-matches `foobar` (classic tmux gotcha). Use the exact-match form `-t =foo` everywhere (`has-session` at `tmux.go:200` and all `-t` uses). Names with `.` or `:` are interpreted as target syntax or mangled by tmux.
- `ls` with no tmux server should behave like `tmux ls` (nothing, consistent exit code), not "Unable to list all tmux sessions".
- Check `exec.LookPath("tmux")` at startup with a clear message.
- Optionally `new-session -d` (do not attach), otherwise attach/switch like `tmux new -A`; today it only prints "Run `tmxu attach ...`" (`cmd.go:550`).
- Document the tested tmux version.

### 7.8 Destructive operations, dry-run, idempotency

- `--dry-run` on restore, `new -templ`, `delete-template`, `save` (print actions instead of doing them; trivial with the Runner).
- Restore without `--force` is already idempotent per session (skips existing, `cmd.go:271-273`); document it. See 1.9 for `--force` and rollback.
- `delete-template` has no confirmation, no `--yes`, and fails with a generic message on a missing template; make it a clear error, or exit 0 with `-f` (`rm` semantics).
- `save-sessions` overwrites the only backup: atomic write (7.9) and keep a `.bak`.

### 7.9 Files, permissions, atomic writes

- Saved sessions contain working-directory paths and window names, so use `0600` files and `0700` dirs (currently `0644`/`0755`, `config.go:36,66,124,212`).
- `os.WriteFile` truncates in place, so a crash or Ctrl-C mid-write corrupts the only backup. One helper:
  ```go
  func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
      f, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
      // write, f.Sync(), f.Close(), os.Chmod(f.Name(), perm), os.Rename(f.Name(), path)
  }
  ```
  Use it for both sessions and templates.
- Add `"version": 1` to the JSON so format changes can migrate cleanly.

### 7.10 Shell completion (hand-written)

- `tmxu completion zsh|bash|fish` prints a script to stdout; embed the scripts with `//go:embed` or generate them with `text/template` from the command registry so they cannot drift.
- Dynamic names need a hidden `tmxu __complete <cmd> <prefix>` subcommand that prints session names (from `tmux list-sessions -F '#{session_name}'`) for `attach`/`save-template`, and template names (strip `.json`) for `delete-template` and `new -templ`. The shell scripts call it.
- Test the completion functions with the fake runner and a temp config dir.
- Install via the formula/release script, with instructions in README and `completion --help`.
- **Library alternative:** cobra ships `completion` with `ValidArgsFunction` and `RegisterFlagCompletionFunc`, so none of the above is needed.

### 7.11 Man page (hand-written)

Write `man/tmxu.1` in roff with sections NAME, SYNOPSIS, DESCRIPTION, COMMANDS, OPTIONS, FILES, ENVIRONMENT, EXIT STATUS, EXAMPLES, SEE ALSO (tmux(1)). Install it from the release script / formula (`man1.install`). `go install` users get no man page, so keep `--help` complete and say so in the README. Keep it in sync with flags by reviewing it whenever a flag changes. **Library alternative:** `cobra/doc` (`doc.GenManTree`) or `go-md2man`.

### 7.12 Other conventions

- Prefix every error `tmxu: ...` so scripts can grep the origin. No trailing spaces in messages (hurts `diff`-based tests and copy-paste).
- Respect `TERM=dumb` and `COLUMNS` for table width; multibyte-safe padding (`output.go:16` and `menu.go:43` use `len`).
- Optional: `ls` shows attached status (`#{session_attached}`) and marks the current session.
- tmux is Unix-only: add `//go:build !windows` or a clear unsupported-OS message.
- Alias naming is confusing (`s` save, `st` save-template, `dt` delete-template); document aliases in help and the man page.

---

## 8. Repo hygiene and release

- **`.gitignore`** (`CLAUDE.md`, `tmxu`): the unanchored `tmxu` also ignores any directory or file named `tmxu` (a future `cmd/tmxu/`). Use `/tmxu`. Add `.DS_Store`, `dist/`, `.claude/`, `*.test`, `coverage.out`. No stray binary is committed.
- **CLAUDE.md** is gitignored and stale: it mentions `gh.go` (the file is `versions.go`), `Formula/tmxu.rb` and `brew install --formula Formula/tmxu.rb` (no `Formula/` directory exists) and `tSession` names. Either track it and fix it, or leave it untracked.
- **README:** Go "1.16+" is wrong (`go.mod` requires 1.26). Add the tmux version, install sections, completion setup, exit codes, JSON format, the fzf recipe, and how it differs from tmux-resurrect, tmuxinator and sesh. Save/restore prompt for confirmation, so cron backups are impossible until `-y` exists.
- **CI** (`.github/workflows/release.yaml`): the only workflow tags `version.txt` on push to main. There is no build, vet, test or lint on PRs. Add `ci.yaml`: `go build ./...`, `go vet ./...`, `gofmt -l` diff check, `go test -race ./...`, staticcheck, `go mod tidy` diff check. Pin actions by SHA (`actions/checkout@v5` is unpinned); add Dependabot for modules and actions.
- **Release flow problems:** `git tag "$VERSION"` fails if the tag already exists (any re-edit of `version.txt`); no `git config user.name/email`; no check that the tag matches `v*`; no GitHub Release, no binaries, no checksums, no changelog. The version lives in two places (`version.txt` embedded in the binary, and the tag).
- **Fix (stdlib-first):** trigger on `on: push: tags: ['v*']`. Use the git tag as the single source of truth: `-ldflags "-s -w -X main.version=$TAG"` with `-trimpath` and `CGO_ENABLED=0`, falling back to `debug.ReadBuildInfo().Main.Version` for `go install` users. Delete `version.txt` and the `paths` trigger. Build a matrix for darwin/linux x amd64/arm64, produce `sha256sum` checksums, and create the GitHub Release with `gh release create`.
- **Library/tool alternative:** goreleaser (builds, archives with LICENSE/README/man/completions, checksums, changelog from the `fix:`/`chore:` commits, Homebrew tap with `generate_completions_from_executable` and `man1.install`).
- **justfile:** `run CMD:` runs `go run ./ CMD` and passes the literal string, not the variable (needs `{{CMD}}`); use `*ARGS` with `{{ARGS}}` to accept several arguments. Mixed tabs and spaces (`test:` uses a tab), trailing whitespace on `go fmt ./... ` and `lint: fmt `. `lint` runs `go fix -diff ./...`, which is not a linter. Add `build`, `install`, `check` (fmt + vet + staticcheck + test -race), put every recipe in a `[group]`, and make `go test` quiet by default.
- **Docs:** add CONTRIBUTING.md (build with `just`, run tests, commit convention matching the `fix:`/`chore:` history, release by tagging), CHANGELOG.md, `.editorconfig`, issue and PR templates. LICENSE exists; check the holder name and year.
- **Dependencies:** `golang.org/x/term v0.39.0` (+ indirect `x/sys`), used only by `menu.go`. Keep it. Run `go mod tidy`.
- **Formula:** no `Formula/` directory exists, so nothing has drifted. Remove the references or add a tap repository.

---

## Suggested order (one Go concept per step)

1. **Fix the real bugs** (section 1) with a failing test first. Concepts: table-driven tests, `time.Duration`, `strconv`.
2. **Error hygiene** (section 2): `%w`, lowercase messages, `errors.Is/As`, sentinels, `Run(args) error`, one `os.Exit`, stderr vs stdout. Concepts: error wrapping, exit codes.
3. **Runner seam and typed parsing** (5.1, 5.2). Concepts: small interfaces, fakes, dependency injection.
4. **Tests on top of the seam** (5.3), including `t.TempDir` and golden files.
5. **Config store and atomic writes** (4, 7.9). Concepts: `os.CreateTemp`, `os.Rename`, `filepath.Join`, generics for `readJSON[T]`.
6. **Deduplicate** (section 4): `snapshotSession`, `restoreSession`, flags and help from one source. Take the schema change only after golden tests pin the old format.
7. **`context` and signals** (6.1, 6.2). Concepts: `exec.CommandContext`, `signal.NotifyContext`, timeouts.
8. **Goroutines** (6.3): the parallel snapshot with `sync.WaitGroup` and an indexed result slice; run under `-race`.
9. **Unix polish** (section 7): help/exit codes/flags first, then TTY and `$TMUX` handling, then `--json` and `-0`, then hand-written completion and the man page last (most work, least idiomatic by hand).
10. **CI and release** (section 8).

Optional shortcut if you later decide that learning speed matters less: cobra (flags, completion, man), `x/mod/semver`, `x/sync/errgroup`, goreleaser.
