package expect

import "testing"

func Eq[T comparable](t *testing.T, got, want T) {
	t.Helper()

	if got != want {
		t.Errorf("\n got: \n\t `%v` \n want: \n\t `%v` \n\n", got, want)
	}
}
