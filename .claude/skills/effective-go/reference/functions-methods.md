# Functions, methods, interfaces

Source: [Functions](https://go.dev/doc/effective_go#functions), [Methods](https://go.dev/doc/effective_go#methods), [Interfaces and other types](https://go.dev/doc/effective_go#interfaces_and_types), [Embedding](https://go.dev/doc/effective_go#embedding).

## Functions

- **Multiple returns:** `(value, error)`, `(value, ok)`. Error is always last.
- **Named results:** use for documentation or for `defer` to modify them. Avoid naked `return` in long functions.
- **`defer`:** runs at function return, LIFO. Put it right after the acquire (`f, err := os.Open(...)`; check err; `defer f.Close()`). Beware `defer` in loops. Arguments are evaluated at the `defer` statement.
- **Small functions:** one job, few params. Group related params into a struct once the list grows.

## Methods

- **Pointer vs value receiver:** pointer if the method mutates, the type is large, or contains a `sync.Mutex`. Be consistent across one type.
- **Value receiver** can be called on values and pointers. Pointer receiver only on addressable values.
- **Do not copy** types that hold locks.

## Interfaces

- **Small interfaces,** often one method. Define them where they are *used* (the consumer), not where they are implemented.
- **Accept interfaces, return concrete types.**
- **Implicit satisfaction:** no `implements`. A compile-time check is `var _ Runner = (*execRunner)(nil)`.
- **Interface for a seam:** wrapping `exec.Command` behind a small interface is what makes code testable without the real binary.
- **Do not export an interface "just in case."** Add it when a second implementation or a test needs it.
- **Type assertions:** `v, ok := x.(T)`. The single-value form panics.
- **Generality:** if a type only exists to implement an interface, export the interface constructor, not the type.

## Embedding

- Embedding promotes methods and fields: composition, not inheritance.
- The embedded type's receiver is the inner value, not the outer one.
- Name conflicts: the shallower depth wins; same depth = compile error on use.

## The blank identifier

`_ = x` to silence unused, `_ , err :=` to ignore a value, `import _ "pkg"` for side effects. Never ignore an `error` silently without a comment.
