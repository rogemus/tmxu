# Control structures and data

Source: [Control structures](https://go.dev/doc/effective_go#control-structures), [Data](https://go.dev/doc/effective_go#data), [Initialization](https://go.dev/doc/effective_go#initialization).

## Control flow

- **No `else` after `return`:** handle the error first, keep the happy path unindented.
- **`if` with init statement:** `if err := f(); err != nil {` keeps scope small.
- **`for` is the only loop:** `for range n` (Go 1.22+), `for i, v := range s`, `for k := range m`.
- **`switch`** has no fall-through, cases can be expressions, and `switch {` replaces if-else chains.
- **Type switch:** `switch v := x.(type)`.
- **Redeclaration:** `:=` may redeclare a variable if at least one new variable is on the left in the same scope. Watch for accidental shadowing of `err`.

## Allocation and data

- **`new(T)`** returns `*T` zeroed. **`make`** is only for slices, maps, channels and returns the value, not a pointer.
- **Make the zero value useful:** `sync.Mutex`, `bytes.Buffer` work without init. Design types the same way.
- **Composite literals:** `&File{fd: fd}`, labeled fields.
- **Slices:** a view over an array (pointer, len, cap). Know when `append` reallocates and when two slices share memory. See https://go.dev/blog/slices-intro.
- **Preallocate** when the size is known: `make([]T, 0, n)`.
- **Maps:** nil map reads are fine, writes panic. Comma-ok idiom `v, ok := m[k]`. Delete with `delete`. Set = `map[T]struct{}`.
- **Iteration order of maps is random.** Sort keys (`slices.Sorted(maps.Keys(m))`) when order matters.
- **Printing:** `%v`, `%+v`, `%#v`, `%T`, `%q`. Implement `String()` to control `%v`.
- **Constants:** `iota` for enumerations, typed constants.
- **`init`:** avoid unless needed. Prefer explicit construction.

## Modern replacements (post Effective Go)

`slices` (`Contains`, `ContainsFunc`, `Index`, `Sort`, `SortFunc`, `Reverse`, `Clone`), `maps` (`Keys`, `Values`, `Clone`), builtins `min`, `max`, `clear`.
