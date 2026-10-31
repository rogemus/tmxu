// Package config holds the app configuration, such as the tmxu version and the
// path where saved sessions live, so it can be passed to commands.
package config

type Cfg struct {
	Path    string
	Version string
}

func New(version, path string) Cfg {
	return Cfg{Path: path, Version: version}
}
