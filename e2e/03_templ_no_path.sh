#!/usr/bin/env bash
# new-session -templ without -path: panes start in the client's directory (a
# new-window without -c uses the client's directory, not the session's).
source "$(dirname "$0")/lib.sh"

fixture_split
(cd "$work/start" && "$bin" new-session -templ=split nopath >/dev/null)

want=$(cd "$work/start" && pwd -P)
check "no -path: every pane starts in the current directory" "$want" "$(paths nopath)"
finish
