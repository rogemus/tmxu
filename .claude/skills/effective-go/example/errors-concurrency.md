# Examples: errors and concurrency

Analogous code on a made-up domain. Not a fix for the user's code.

## Wrap and inspect

```go
var ErrNoRecord = errors.New("no record")

func find(id string) (Record, error) {
	// ... returns fmt.Errorf("find %q: %w", id, ErrNoRecord) when missing
}

if errors.Is(err, ErrNoRecord) {
	// handle
}
```

Read: https://go.dev/blog/go1.13-errors. Guiding question: why does `err == ErrNoRecord` stop working after wrapping?

## Typed error with fields

```go
type RunError struct {
	Args   []string
	Stderr string
	Err    error
}

func (e *RunError) Error() string { /* ... */ }
func (e *RunError) Unwrap() error { return e.Err }

var re *RunError
if errors.As(err, &re) {
	// use re.Stderr
}
```

Fill in `Error()` yourself. What should it print so the log line is useful?

## Cancellation

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

cmd := exec.CommandContext(ctx, "sleep", "30")
```

Guiding question: which exit code does a shell report after Ctrl-C, and who should call `os.Exit`?

## Exercise gap

```go
sem := make(chan struct{}, 4)
var wg sync.WaitGroup
for _, job := range jobs {
	// start the goroutine: acquire, release, wait group, result slot
}
wg.Wait()
```

Where do the results go so there is no data race? Run it with `-race` to check.
