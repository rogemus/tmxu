#!/usr/bin/env bash
# new-session -templ -path=<abs>: every pane starts in the given directory.
source "$(dirname "$0")/lib.sh"

fixture_split
(cd "$work/start" && "$bin" new-session -templ=split -path="$work/abs" abspath >/dev/null)

want=$(cd "$work/abs" && pwd -P)
check "-path ABS: every pane starts in the given directory" "$want" "$(paths abspath)"
finish
