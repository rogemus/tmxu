#!/usr/bin/env bash
# new-session -templ -path=../: a relative path is resolved against the
# client's directory.
source "$(dirname "$0")/lib.sh"

fixture_split
(cd "$work/start/sub" && "$bin" new-session -templ=split -path=../ relpath >/dev/null)

want=$(cd "$work/start" && pwd -P)
check "-path ../: every pane starts in the parent directory" "$want" "$(paths relpath)"
finish
