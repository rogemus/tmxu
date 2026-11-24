#!/usr/bin/env bash
# Runs every e2e/NN_*.sh check, each in its own isolated tmux server.
#
# Usage: just e2e            (all checks)
#        just e2e 03         (checks whose file name starts with 03)
#        bash e2e/03_templ_no_path.sh   (one file directly)

set -euo pipefail

dir=$(cd "$(dirname "$0")" && pwd)
root=$(cd "$dir/.." && pwd)

command -v tmux >/dev/null || { echo "e2e: tmux not found in PATH" >&2; exit 2; }

build=$(mktemp -d /tmp/tmxu-e2e-bin.XXXXXX)
trap 'rm -rf "$build"' EXIT
(cd "$root" && go build -o "$build/tmxu" .)
export TMXU_E2E_BIN="$build/tmxu"

failed=0
ran=0
prefix=${1:-[0-9]} # default: any NN_ file; unquoted below so the glob expands
for t in "$dir"/$prefix*_*.sh; do
  [[ -e "$t" ]] || continue
  ran=$((ran + 1))
  echo "== $(basename "$t")"
  bash "$t" || failed=$((failed + 1))
done

echo
if ((ran == 0)); then
  echo "e2e: no check matches '${1:-}'" >&2
  exit 2
fi
if ((failed > 0)); then
  echo "e2e: $failed of $ran file(s) failed" >&2
  exit 1
fi
echo "e2e: all $ran file(s) passed"
