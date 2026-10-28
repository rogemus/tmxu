# Restore performance (parked)

Not urgent. Parked for later; nothing here is implemented.

## Symptom

After `tmxu restore`, attaching to a restored session shows panes "catching up": prompts redraw and shells are still starting. Preferably the user sees panes only once they are ready, even if restore takes a bit longer.

## Measurement

- `time zsh -i -c exit`: 0.127 s total, 0.03 s user + 0.03 s system (about 60 ms CPU). A single shell starts quickly, so the user's zshrc is not the problem. Note that `-c exit` skips prompt hooks (precmd, prompt frameworks), so a real pane costs somewhat more.
- The cost is aggregate: a real setup restores 15 sessions, 58 windows and 59 panes at once, so about 59 shells x ~60 ms CPU = ~3.5 s of CPU contending across cores. The user attaches while many are still booting.
- Not yet measured: `time tmxu restore` (answer the prompt). Do this first to get a baseline. If it is under about 2 s, steps 1 and 2 below should be enough.

## Likely causes (from reading `cli/cmd.go` restore loop and `cli/tmux.go`; not profiled)

1. **Detached size, then resize on attach.** `new-session -d` creates the session at tmux's default 80x24, and `select-layout` is applied at that size. Attaching resizes to the real terminal, so every pane gets SIGWINCH and the shell redraws its prompt.
2. **Shells still booting.** Each pane spawns the default shell (`/bin/zsh`) as soon as it is created. Restore returns immediately, so the user attaches during startup.
3. **Many forks.** Each tmux call is a separate `exec.Command`: about 4 per single-pane window plus `has-session` and `new-session`, several hundred for 59 panes. Real but small compared to the shells.

## Ideas, ordered by value

1. **Wait for panes to be ready before returning (main fix).**
   - After creating all panes, poll each pane until its prompt has drawn: `tmux capture-pane -p -t <pane>` returns non-empty, every ~50 ms, ~5 s timeout per session.
   - Poll sessions in parallel (one goroutine per session, `sync.WaitGroup`), so the total wait is the slowest shell, not the sum.
   - Print `waiting for shells...` to stderr; on timeout warn and continue.
   - Add a `-no-wait` flag to skip it.
   - Good goroutine exercise. The polling is read-only.
2. **Create sessions at the real size.** Add `-x <cols> -y <rows>` to `new-session`, taken from `term.GetSize(int(os.Stdout.Fd()))` (`golang.org/x/term` is already a dependency). Fall back to no flags when stdout is not a TTY. Removes the SIGWINCH redraw on attach when the attaching terminal has the same size.
3. **Batch tmux calls (low priority).** Combine a window's `new-window`, `split-window`, `select-pane -T` and `select-layout` into one `tmux a \; b \; c` invocation. Also skip `select-layout` for single-pane windows (see `TODO.md`, section 4).
4. **Staggered start (optional).** Restore in batches of 3-4 sessions and wait for each batch before the next, to cap CPU contention at the cost of a longer total restore. Only worth it if waiting for everything still feels janky.

Alternative readiness signal: `send-keys 'tmux wait-for -S <name>' Enter` per pane, then `tmux wait-for <name>`. A shell reads queued input only after startup finishes, so this is an exact signal, but the command echoes visibly in the pane. Prefer `capture-pane` polling.

## Open decisions

- Wait by default (with `-no-wait`), or opt-in (`-wait`)? Recommendation: default.
- A shell with an empty prompt never produces output, so it hits the timeout (5 s per session, in parallel).
- If the attaching terminal has a different size than the one used at restore time, the resize still happens; step 2 only helps when sizes match.

## Files that would change

- `cli/tmux.go`: size flags in `NewSession`, a new `WaitPaneReady`, optional batching.
- `cli/cmd.go`: wait step and `-no-wait` flag in `restore-sessions` and `new-session -templ`.
