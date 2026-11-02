# Examples: functions, methods, interfaces

Analogous code on a made-up domain. Not a fix for the user's code.

## defer right after acquire

```go
f, err := os.Open(name)
if err != nil {
	return err
}
defer f.Close()
```

Guiding question: what happens to the `defer` if you put it inside a `for` loop over many files?

## Consumer-side small interface

```go
// lives in the package that CALLS it
type fetcher interface {
	Fetch(ctx context.Context, id string) ([]byte, error)
}

func Sync(ctx context.Context, f fetcher, ids []string) error { /* ... */ }
```

Guiding question: who defines the interface here, and what does a test pass in instead of the real client?

## Compile-time check

```go
var _ fetcher = (*httpFetcher)(nil)
```

## Receiver choice

```go
type Counter struct {
	mu sync.Mutex
	n  int
}

func (c ??? Counter) Inc() { /* locks mu, n++ */ }
```

What goes in `???`, and what does `go vet` say if you pick the other one?
