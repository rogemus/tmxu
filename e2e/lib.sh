#!/usr/bin/env bash
# Shared helpers for the e2e checks. Source it, do not run it:
#   source "$(dirname "$0")/lib.sh"
#
# Every check runs on an isolated tmux server (private TMUX_TMPDIR) with a
# private config dir (XDG_CONFIG_HOME), so your own sessions and templates are
# never touched. The work dir lives under /tmp on purpose: a tmux socket path
# is limited to ~104 bytes, and $TMPDIR on macOS is too long.

set -euo pipefail

e2e_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

command -v tmux >/dev/null || { echo "e2e: tmux not found in PATH" >&2; exit 2; }

work=$(mktemp -d /tmp/tmxu-e2e.XXXXXX)
cleanup() {
  tmux kill-server 2>/dev/null || true
  rm -rf "$work"
}
trap cleanup EXIT

export TMUX_TMPDIR="$work/tmux" XDG_CONFIG_HOME="$work/xdg"
unset TMUX
mkdir -p "$TMUX_TMPDIR" "$XDG_CONFIG_HOME" "$work/start/sub" "$work/abs"

# run.sh builds once and exports TMXU_E2E_BIN; a single check builds its own.
if [[ -n "${TMXU_E2E_BIN:-}" ]]; then
  bin=$TMXU_E2E_BIN
else
  bin="$work/tmxu"
  (cd "$e2e_root" && go build -o "$bin" .)
fi

failures=0
pass() { printf 'ok    %s\n' "$1"; }
fail() { printf 'FAIL  %s\n' "$1" >&2; failures=$((failures + 1)); }

# check NAME EXPECTED ACTUAL
check() {
  if [[ "$2" == "$3" ]]; then
    pass "$1"
  else
    fail "$1"
    printf '      want: %s\n      got:  %s\n' "$2" "$3" >&2
  fi
}

# Call last; exits 1 when any check failed.
finish() {
  if ((failures > 0)); then
    echo "e2e: $failures check(s) failed" >&2
    exit 1
  fi
}

# Window layouts of session $1 with the volatile pane ids ("I":"%N") removed.
layouts() {
  tmux list-windows -t "$1:" -F '#{window_index} #{window_layout}' |
    sed -E 's/,"I":"%[0-9]+"//g'
}

# Pane sizes of session $1.
sizes() {
  tmux list-panes -s -t "$1:" -F '#{window_index}.#{pane_index} #{pane_width}x#{pane_height}'
}

# Distinct pane start directories of session $1.
paths() {
  tmux list-panes -s -t "$1:" -F '#{pane_current_path}' | sort -u
}

# Session "src": two windows with three panes each, the second one with a
# non-default layout; saved as the template "split".
fixture_split() {
  (cd / && tmux new-session -d -s src -x 200 -y 50 -c /)
  tmux split-window -h -t 'src:' -c /
  tmux split-window -v -t 'src:' -c /
  tmux new-window -t 'src:' -c /
  tmux split-window -h -t 'src:' -c /
  tmux split-window -v -t 'src:' -c /
  tmux select-layout -t 'src:' main-vertical

  "$bin" save-template -name=split src >/dev/null
}
