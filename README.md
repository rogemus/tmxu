# tmxu

Save and restore tmux sessions, and build new ones from templates. One small Go binary, plain JSON files.

## What is tmxu?

`tmxu` does two things:

1. **Session persistence**: save every running session (windows, panes, start directories, layouts) and restore them later.
2. **Templates**: turn a running session into a reusable blueprint and create new sessions from it under any name.

## Requirements

- **tmux 3.8+** in `PATH`. tmxu reads window layouts as JSON, which tmux added in 3.8.
- **Go 1.26+** to build from source.

## Installation

```bash
go install github.com/rogemus/tmxu@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`. From a clone, `just install` does the same.

## Quick start

```bash
tmxu new my-project            # create a detached session
tmxu ls                        # list sessions
tmxu attach my-project         # attach to it
tmxu save                      # save all sessions
tmxu save-template my-project  # save one session as a template
```

## Commands

Flags go **before** the session or template name: `tmxu new -path /tmp/app name`.

### Sessions

| Command                 | Aliases       | Description                                           |
| ----------------------- | ------------- | ----------------------------------------------------- |
| `list-sessions`         | `list`, `ls`  | List running sessions                                 |
| `attach-session <name>` | `attach`, `a` | Attach to a running session                           |
| `new-session <name>`    | `new`, `ns`   | Create a detached session, optionally from a template |
| `dump`                  | `d`           | Print running sessions, windows and panes             |

| Command          | Flag              | Description                                                |
| ---------------- | ----------------- | ---------------------------------------------------------- |
| `list-sessions`  | `-json`           | Print JSON                                                 |
| `attach-session` | `-menu`           | Pick the session from an interactive menu                  |
| `new-session`    | `-path <dir>`     | Start directory for all panes (default: current directory) |
| `new-session`    | `-templ <name>`   | Build the session from a template                          |
| `dump`           | `-json`           | Print the JSON that `save` writes, layouts included        |
| `dump`           | `-session <name>` | Dump one session instead of all                            |

`dump` prints a tree with the pane names and start directories. It saves nothing.

### Save and restore

| Command            | Aliases        | Description                                    |
| ------------------ | -------------- | ---------------------------------------------- |
| `save-sessions`    | `save`, `s`    | Save all running sessions to the sessions file |
| `restore-sessions` | `restore`, `r` | Recreate the saved sessions                    |

| Command            | Flag     | Description                                              |
| ------------------ | -------- | -------------------------------------------------------- |
| `save-sessions`    | `-yes`   | Skip the confirmation prompt                             |
| `restore-sessions` | `-force` | Replace sessions that already exist (default: skip them) |

### Templates

| Command                   | Aliases       | Description                          |
| ------------------------- | ------------- | ------------------------------------ |
| `list-templates`          | `lt`          | List saved templates                 |
| `save-template <session>` | `st`          | Save a running session as a template |
| `show-template <name>`    | `show`, `sht` | Show a template as a tree            |
| `delete-template <name>`  | `dt`          | Delete a template (no confirmation)  |

| Command          | Flag           | Description                                              |
| ---------------- | -------------- | -------------------------------------------------------- |
| `list-templates` | `-json`        | Print JSON                                               |
| `save-template`  | `-name <name>` | Template name (default: the session name)                |
| `save-template`  | `-force`       | Replace a template that already exists (default: refuse) |
| `show-template`  | `-json`        | Print the saved template JSON                            |

A template keeps windows, panes and layouts but no session name and no pane paths, so it can be reused under any name. `save-template` refuses to replace an existing template unless you pass `-force`. Template names are turned into file names (`My Proj` and `my-proj` are the same file).

### Utility

| Command          | Aliases | Description                       |
| ---------------- | ------- | --------------------------------- |
| `paths`          |         | Print where tmxu keeps its files  |
| `version`        | `v`     | Print the tmxu version            |
| `help [command]` |         | Show help for tmxu or one command |

`paths` and `version` accept `-json`.

## Usage examples

```bash
# Save and restore everything
tmxu save -yes                        # no prompt
tmxu restore                          # recreate saved sessions, skip existing ones
tmxu restore -force                   # replace existing ones

# Templates
tmxu save-template -name web api      # session "api" -> template "web"
tmxu show-template web                # look at it
tmxu new -templ web -path ~/code/shop shop

# Scripting
tmxu ls -json
tmxu dump -session api -json
tmxu paths -json
```

## Files

```
~/.config/tmxu/                 # or $XDG_CONFIG_HOME/tmxu
├── tmux-sessions.json          # written by `save`
└── templates/
    ├── web.json
    └── api.json
```

`tmxu paths` prints the real locations. Files are written with mode `0600`, and every file is JSON you can read, edit or keep in version control.

## JSON format

`save`, `dump -json` and every template use the same shape:

```json
{
  "version": 1,
  "sessions": [
    {
      "session_name": "api",
      "session_windows": [
        {
          "window_name": "editor",
          "window_index": 1,
          "window_panes": [{ "pane_index": 1, "pane_path": "/home/me/api" }]
        },
        {
          "window_name": "dev",
          "window_index": 2,
          "window_layout": { "V": 2, "L": { "t": "h", "c": [] } },
          "window_panes": [
            { "pane_index": 1, "pane_path": "/home/me/api" },
            { "pane_index": 2, "pane_path": "/home/me/api" }
          ]
        }
      ]
    }
  ]
}
```

- `window_layout` is the raw tmux 3.8 layout. It is stored only for windows with more than one pane (the example above is shortened).
- Pane names are not stored. They are derived as `<window name>_<pane index>`.
- A template is the same file with exactly one session. Its `session_name` and every `pane_path` are empty.
- tmxu restores panes by their saved index, so the restoring tmux should use the same `base-index` and `pane-base-index` as the one that saved them.

Other commands print small one-line JSON with `-json`:

| Command          | Output                                                 |
| ---------------- | ------------------------------------------------------ |
| `ls`             | `{"sessions":[{"session_name":"api"}]}`                |
| `list-templates` | `{"templates":["web"]}`, `{"templates":[]}` when empty |
| `paths`          | `{"config":"...","templates":"...","sessions":"..."}`  |
| `version`        | `{"version":"..."}`                                    |

## Exit codes

| Code | Meaning                                                      |
| ---- | ------------------------------------------------------------ |
| `0`  | Success                                                      |
| `1`  | Runtime error (tmux failed, file missing, session not found) |
| `2`  | Usage error (unknown command or flag, missing argument)      |

Errors are printed to stderr as `tmxu: <message>`. Data and success messages go to stdout, so `tmxu ls -json | jq` stays clean.

## How it differs

- **tmux-resurrect**: a tmux plugin (installed with tpm, driven by key bindings) that also restores pane contents and some running programs. tmxu is a standalone binary that saves layout and start directories only, and adds templates and `-json` output.
- **tmuxinator**: you write YAML project files by hand. tmxu captures a live session into a template, so you arrange the windows once in tmux and reuse them.
- **sesh**: a session switcher built around zoxide and fzf. tmxu does not pick or jump between sessions; it focuses on save, restore and templates.

## Tips

- Templates and the sessions file are plain JSON: keep them in version control.
- Use `save` / `restore` for the exact state of your workspace, and templates for the shape of a project.
- Run `tmxu dump` before `save` to see what will be written.
