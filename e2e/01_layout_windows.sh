#!/usr/bin/env bash
# select-layout: raw tmux JSON with stale pane ids is accepted and applied, so
# every restored window has the layout of the saved one.
source "$(dirname "$0")/lib.sh"

fixture_split
(cd "$work/start" && "$bin" new-session -templ=split layout >/dev/null)

check "layout: window layouts match source" "$(layouts src)" "$(layouts layout)"
finish
