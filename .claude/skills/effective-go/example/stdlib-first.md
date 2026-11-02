# Examples: stdlib in place of a package

Analogous code on a made-up domain. Not a fix for the user's code.

## Subcommands with `flag`

```go
fs := flag.NewFlagSet("export", flag.ContinueOnError)
out := fs.String("o", "-", "output file")
if err := fs.Parse(args); err != nil {
	// what exit code do usage errors conventionally use?
}
```

Read: https://pkg.go.dev/flag. Replaces cobra for a small tool.

## Logging with `slog`

```go
log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
log.Debug("run", "args", args)
```

Guiding question: why stderr and not stdout?

## Slice helpers

```go
idx := slices.IndexFunc(items, func(i Item) bool { return i.Name == want })
// what does it return when nothing matches, and how do you check for that?
```

## Atomic file write (gap)

```go
// 1. create temp file in the same directory: os.CreateTemp
// 2. write, Sync, Close
// 3. ??? to replace the target atomically
```

Which `os` function fills step 3, and why must the temp file be on the same filesystem?
