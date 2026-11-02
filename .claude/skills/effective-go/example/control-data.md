# Examples: control flow and data

Analogous code on a made-up domain. Not a fix for the user's code.

## Happy path unindented

```go
func load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	// ... continue without an else
}
```

Guiding question: where in your function does the success path get indented?

## Set and ordered keys

```go
seen := map[string]struct{}{}
// add: seen[name] = struct{}{}
// need sorted output: which functions in `maps` and `slices` chain here?
```

See https://pkg.go.dev/maps and https://pkg.go.dev/slices.

## Slice aliasing

```go
a := []int{1, 2, 3, 4}
b := a[:2]
b = append(b, 99)
fmt.Println(a) // predict the output before running it
```

Read: https://go.dev/blog/slices-intro. What changes if you write `a[:2:2]`?

## Exercise gap

```go
var names []string
for _, p := range people {
	names = append(names, p.Name)
}
// People count is known. What one-line change avoids repeated reallocation?
```
