# Formatting and naming

Source: [Formatting](https://go.dev/doc/effective_go#formatting), [Commentary](https://go.dev/doc/effective_go#commentary), [Names](https://go.dev/doc/effective_go#names).

## Check when reviewing

- **gofmt:** all code is `go fmt` clean. No style debates beyond that.
- **Doc comments:** every exported identifier has a comment starting with its name. A package has one package comment. See https://go.dev/doc/comment.
- **MixedCaps:** `userName`, `UserName`. No underscores in Go names (file names may use them).
- **Exported by case:** capital first letter = exported. Unexport whatever is not API.
- **Package names:** short, lowercase, single word, no `util`/`common`/`helpers` grab-bags. The package name is part of the call site: `bytes.Buffer`, not `bytes.BytesBuffer`. See https://go.dev/blog/package-names.
- **Getters:** field `owner` gets `Owner()`, not `GetOwner()`. Setter is `SetOwner()`.
- **Interface names:** one-method interface = method name + `-er` (`Reader`, `Writer`, `Stringer`).
- **Short scope, short name:** `i`, `r`, `err` in a few lines; longer names for wide scope.
- **Canonical method names:** `String`, `Error`, `Read`, `Write`, `Close` keep their standard meaning and signature.
- **Semicolons and braces:** opening brace on the same line.

## Questions to ask the user

- What does the caller read at the call site? Does the package name repeat?
- Is this identifier part of the API, or can it be unexported?
- Would a comment starting with the name pass `go doc`?
