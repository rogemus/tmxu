# Errors and concurrency

Source: [Errors](https://go.dev/doc/effective_go#errors), [Concurrency](https://go.dev/doc/effective_go#concurrency), https://go.dev/blog/go1.13-errors.

## Errors

- **Errors are values.** Return them, do not hide them. Check every one.
- **Add context, keep the cause:** `fmt.Errorf("load %s: %w", path, err)`. Use `%w` once per call (multiple `%w` is valid since 1.20).
- **Inspect with `errors.Is` / `errors.As`,** never string compare or `==` on wrapped errors.
- **Sentinels:** `var ErrNotFound = errors.New("not found")`, exported only if callers branch on them.
- **Typed errors:** a struct implementing `Error() string` (and `Unwrap()`) when callers need fields (exit code, stderr).
- **Combine:** `errors.Join` collects several errors.
- **Messages:** lowercase, no trailing punctuation, no "failed to" stacking.
- **File checks:** `errors.Is(err, fs.ErrNotExist)` instead of `os.Stat` then `os.IsNotExist`.
- **`panic`** is for programmer error or unrecoverable state, never for expected failures. Libraries return errors. `recover` only at a boundary (e.g. goroutine top level).
- **Process exit:** only `main` calls `os.Exit`. Errors go to stderr, data to stdout.

## Concurrency

- **Share memory by communicating.** Channels pass ownership. A mutex is fine for simple shared state.
- **Goroutine lifetime:** every goroutine must have a clear way to stop. Know who waits for it (`sync.WaitGroup`) and what cancels it (`context`).
- **Channels:** unbuffered = rendezvous, buffered = bounded queue. Sender closes, receiver never closes. A `nil` channel blocks forever.
- **Semaphore** with a buffered channel: `sem := make(chan struct{}, n)`.
- **`select`** with `ctx.Done()` for cancellation and timeouts.
- **Loop variables:** per-iteration since Go 1.22, so the old `v := v` copy is unnecessary.
- **Data races:** run tests with `-race`. See https://go.dev/doc/articles/race_detector.
- **`context`:** first parameter, named `ctx`, never stored in a struct. `exec.CommandContext` for subprocesses. `signal.NotifyContext` in `main`. See https://go.dev/blog/context.
- **Do not start concurrency before measuring.** Sequential code is easier to review.

## Questions to ask the user

- Who owns this goroutine, and what stops it?
- What does the caller need to know to react: a sentinel, a type, or just text?
- Is this a bug (panic) or an expected condition (error)?
