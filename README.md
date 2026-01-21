# tmxu

Tmux utilities for managing Tmux sessions with ease.

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

| Command | Aliases | Args | Description |
|---------|---------|------|-------------|
| new-session | new, ns | [sessionName] | Create new session based on template |
| attach-session | attach, a | [sessionName] | Attach to running tmux session |
| list-sessions | list, ls | | List all active sessions in tmux |
| save-sessions | save, s | | Save tmux sessions |
| restore-sessions | restore, r | | Restore tmux sessions |
| list-templates | lt | | List all saved templates |
| save-template | st | [sessionName] | Save session as template |
| delete-template | dt | [templateName] | Delete saved template |
| version | v | | Display app version information |
| help | | [command] | Display help information |

Use `tmxu help [command]` to get detailed information about a specific command.

## Templates

Templates allow you to save a tmux session layout as a reusable blueprint. This is useful for creating consistent development environments across different projects.

Unlike `save`/`restore` which captures your entire workspace state, templates store only the session structure (windows and panes) with configurable paths. You can then spawn new sessions from a template with a custom base directory.

Templates are stored as JSON files in `~/.config/tmxu/templates/`.

### Usage

```bash
# Save current session "dev" as a template
tmxu save-template dev
tmxu st dev

# Save with a custom template name
tmxu save-template -name mytemplate dev

# List available templates
tmxu list-templates
tmxu lt

# Create a new session from a template
tmxu new-session -templ dev mysession
tmxu new -templ dev mysession

# Create a new session from a template with custom path
tmxu new-session -path /projects/myapp -templ dev mysession

# Delete a template
tmxu delete-template dev
tmxu dt dev
```
