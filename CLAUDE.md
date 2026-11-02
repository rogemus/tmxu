# CLAUDE.md

## Mode: review only

Learning project. The user writes the code, Claude reviews and teaches.

- Never edit files unless explicitly asked. `README.md` and `TODO.md` may be updated without asking.
- Do not hand over finished fixes. Give doc links, concepts and partial examples on different code.
- Stdlib first. Only `golang.org/x/term` is an allowed dependency.
- Rules and skill: `.claude/rules/go-review-mode.md`, `.claude/skills/effective-go/`.

## Project

`tmxu` is a Go CLI that saves and restores tmux sessions (windows, panes, layouts) and creates sessions from templates. Targets tmux 3.8+ and Go 1.26 (see `go.mod`).

## Commands

```bash
just run <command>      # go run ./ <command>
just test               # go test ./... -v
just lint               # go fmt, go vet, go fix -diff
go build -o tmxu .
```

## Layout

- `main.go`: entry point
- `internal/cli`: command registry, dispatch, help
- `internal/cmd`: one file per command
- `internal/tmux`: tmux calls and the session/window/pane model
- `internal/config`: config directory resolution
- `internal/utils`: small helpers (quoting, confirm, contains)
- `cli/`: old package, removed once the port to `internal/` is done

## Where things are tracked

- `TODO.md`: remaining work, decisions, suggested order
- `AI-REVIEW.md`: reasoning behind each TODO item
- `internal/` has no tests. The user writes them by hand (`TODO.md` section 7).
