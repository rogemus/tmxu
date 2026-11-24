#!/usr/bin/env bash
# The restored panes have the same sizes as the saved ones (proves that the
# layout was applied, not just that the panes were split by default).
source "$(dirname "$0")/lib.sh"

fixture_split
(cd "$work/start" && "$bin" new-session -templ=split layout >/dev/null)

check "layout: pane sizes match source" "$(sizes src)" "$(sizes layout)"
finish
