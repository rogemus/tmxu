# tmxu

Personal Tmux utilities for managing Tmux sessions with ease.

## About

`tmxu` is a lightweight command-line tool written in Go that simplifies tmux session management. It provides convenient commands for attaching to sessions, saving your current workspace configuration, and restoring sessions later.

Saved sessions are stored in `~/.config/tmxu/tmux-sessions.json`, making it easy to backup and restore your tmux workspace across different machines or after system restarts.

## Installation

Install using Go:

```bash
go install github.com/rogemus/tmxu@latest
```

Make sure your `$GOPATH/bin` is in your `PATH` to run the installed binary.

## Available Commands

```
attach [name]              Attach to running tmux session
list                       List all active sessions in tmux
save                       Save tmux sessions
restore                    Restore tmux sessions
list-templates             List all saved templates
save-template [session]    Save session as template
delete-template [name]     Delete saved template
version                    Display app version information
help [command]             Display help information
```

Use `tmxu help [command]` to get detailed information about a specific command.

Templates are stored in `~/.config/tmxu/templates/`.
