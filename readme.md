# tmxu

Personal Tmux utilities for managing Tmux sessions with ease.

## About

tmxu is a lightweight command-line tool written in Go that simplifies tmux session management. It provides convenient commands for attaching to sessions, saving your current workspace configuration, and restoring sessions later.

## Installation

Install using Go:

```bash
go install github.com/rogemus/tmxu@latest
```

Make sure your `$GOPATH/bin` is in your `PATH` to run the installed binary.

## Available Commands

```bash
attach [NAME]    Attach to running tmux session
list             List all active sessions in tmux
save             Save tmux sessions
restore          Restore tmux sessions
version          Display app version information
help             Display help information
```
