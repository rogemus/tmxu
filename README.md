# tmxu

A command-line tool for managing tmux sessions and templates with ease.

## What is tmxu?

`tmxu` is a lightweight Go utility for tmux session management with two core features:

1. **Session Persistence**: Save and restore complete tmux workspaces
2. **Templates**: Create reusable session blueprints for different projects

## Installation

```bash
go install github.com/rogemus/tmxu@latest
```

Make sure `$GOPATH/bin` is in your `PATH`.

## Quick Start

```bash
tmxu new my-project           # Create a new session
tmxu ls                        # List all sessions
tmxu attach my-project         # Attach to session
tmxu save                      # Save all sessions
tmxu save-template my-project  # Save as template
```

## Commands

### Session Management

| Command                 | Aliases       | Description                                   |
| ----------------------- | ------------- | --------------------------------------------- |
| `list-sessions`         | `list`, `ls`  | List all active tmux sessions                 |
| `attach-session [name]` | `attach`, `a` | Attach to a running session                   |
| `new-session [name]`    | `new`, `ns`   | Create new session (optionally from template) |

**attach-session flags:**

- `-menu` - Interactive menu for session selection

**new-session flags:**

- `-path` - Initial path for all panes (default: current directory)
- `-templ` - Template to base session on

### Session Persistence

| Command            | Aliases        | Description                                              |
| ------------------ | -------------- | -------------------------------------------------------- |
| `save-sessions`    | `save`, `s`    | Save all sessions to `~/.config/tmxu/tmux-sessions.json` |
| `restore-sessions` | `restore`, `r` | Restore sessions from saved file                         |

**restore-sessions flags:**

- `-force` - Override existing sessions (use with caution)

### Templates

| Command                   | Aliases | Description              |
| ------------------------- | ------- | ------------------------ |
| `list-templates`          | `lt`    | List all saved templates |
| `save-template [session]` | `st`    | Save session as template |
| `delete-template [name]`  | `dt`    | Delete a template        |

**save-template flags:**

- `-name` - Custom template name (default: session name)

**Templates are stored in:** `~/.config/tmxu/templates/`

### Utility

| Command          | Aliases | Description                        |
| ---------------- | ------- | ---------------------------------- |
| `version`        | `v`     | Show version and check for updates |
| `help [command]` |         | Display help information           |

## Usage Examples

```bash
# Session management
tmxu attach -menu                    # Interactive session picker
tmxu new -templ dev-setup api-proj   # Create from template

# Save/restore workflow
tmxu save                            # Backup all sessions
tmxu restore                         # Restore all sessions
tmxu restore -force                  # Force restore (kills existing)

# Template workflow
tmxu save-template -name mytemplate mysession
tmxu new -templ mytemplate -path ~/projects/app newproject
```

## File Storage

```
~/.config/tmxu/
├── tmux-sessions.json       # Saved sessions
└── templates/               # Template files
    ├── dev-template.json
    └── web-template.json
```

All files are JSON and can be version controlled or manually edited.

## Tips

- **Backup templates**: Templates are just JSON files - version control them!
- **Use templates for structure**: Use save/restore for exact state
- **Interactive menu**: `tmxu attach -menu` for quick switching
- **Path flexibility**: Templates can use any working directory with `-path`

## Requirements

- **tmux** installed and in PATH
- **Go** 1.16+ (for building from source)
