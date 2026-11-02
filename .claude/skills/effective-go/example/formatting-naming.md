# Examples: formatting and naming

Analogous code on a made-up domain. Not a fix for the user's code. Leave gaps for the user.

## Doc comment and getter

```go
// Playlist is an ordered list of tracks.
type Playlist struct {
	owner  string
	tracks []Track
}

// Owner returns the playlist owner.   <- starts with the name, no "Get"
func (p *Playlist) Owner() string { return p.owner }
```

Guiding question: what would `go doc` print for your type today?

## Package name at the call site

```go
// reads badly: stringutil.StringUtilJoin(...)
// reads well:  strs.Join(...)
```

Guiding question: say the call site out loud. Does the package name repeat?

## Exercise gap

```go
type ??? interface {      // one method: Save(...)
	Save(ctx context.Context, p Playlist) error
}
```

What name does the `-er` convention give this interface?
