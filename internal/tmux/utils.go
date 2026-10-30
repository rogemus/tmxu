package tmux

type Comperable interface {
	Eq(el Comperable) bool
}

func Contain[T Comperable](slice []T, el T) bool {
	for _, e := range slice {
		if e.Eq(el) {
			return true
		}
	}

	return false
}
