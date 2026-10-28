package cli

import "testing"

func TestNewSemVer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  semVer
	}{
		{
			name:  "plain version is parsed",
			input: "1.2.3",
			want:  semVer{original: "1.2.3", major: 1, minor: 2, patch: 3},
		},
		{
			name:  "v prefix is stripped from major",
			input: "v1.2.3",
			want:  semVer{original: "v1.2.3", major: 1, minor: 2, patch: 3},
		},
		{
			name:  "zero version is parsed",
			input: "0.0.0",
			want:  semVer{original: "0.0.0"},
		},
		{
			name:  "multi digit parts are parsed",
			input: "v10.200.3000",
			want:  semVer{original: "v10.200.3000", major: 10, minor: 200, patch: 3000},
		},
		{
			name:  "missing patch defaults to zero",
			input: "1.2",
			want:  semVer{original: "1.2", major: 1, minor: 2},
		},
		{
			name:  "missing minor and patch default to zero",
			input: "v3",
			want:  semVer{original: "v3", major: 3},
		},
		{
			name:  "surrounding whitespace is trimmed from original",
			input: "  v1.2.3\n",
			want:  semVer{original: "v1.2.3", major: 1, minor: 2, patch: 3},
		},
		{
			name:  "extra parts are ignored",
			input: "1.2.3.4",
			want:  semVer{original: "1.2.3.4", major: 1, minor: 2, patch: 3},
		},
		{
			name:  "empty string returns zero version",
			input: "",
			want:  semVer{},
		},
		{
			name:  "non numeric input returns zero version",
			input: "abc",
			want:  semVer{original: "abc"},
		},
		{
			name:  "non numeric part defaults to zero",
			input: "1.x.3",
			want:  semVer{original: "1.x.3", major: 1, patch: 3},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := newSemVer(test.input)

			if got != test.want {
				t.Errorf("newSemVer(%q) = %+v, want %+v", test.input, got, test.want)
			}
		})
	}
}

func TestSortSemVer(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "nil slice stays empty",
			input: nil,
			want:  nil,
		},
		{
			name:  "single element is unchanged",
			input: []string{"1.0.0"},
			want:  []string{"1.0.0"},
		},
		{
			name:  "already sorted stays sorted",
			input: []string{"1.0.0", "1.0.1", "1.1.0", "2.0.0"},
			want:  []string{"1.0.0", "1.0.1", "1.1.0", "2.0.0"},
		},
		{
			name:  "reversed is sorted ascending",
			input: []string{"2.0.0", "1.1.0", "1.0.1", "1.0.0"},
			want:  []string{"1.0.0", "1.0.1", "1.1.0", "2.0.0"},
		},
		{
			name:  "major takes precedence over minor and patch",
			input: []string{"2.0.0", "1.9.9"},
			want:  []string{"1.9.9", "2.0.0"},
		},
		{
			name:  "minor takes precedence over patch",
			input: []string{"1.2.0", "1.1.9"},
			want:  []string{"1.1.9", "1.2.0"},
		},
		{
			name:  "numeric order is used instead of lexical",
			input: []string{"v1.10.0", "v1.2.0", "v1.9.0"},
			want:  []string{"v1.2.0", "v1.9.0", "v1.10.0"},
		},
		{
			name:  "v prefixed versions are sorted",
			input: []string{"v0.7.0", "v0.10.0", "v0.6.1"},
			want:  []string{"v0.6.1", "v0.7.0", "v0.10.0"},
		},
		{
			name:  "missing parts sort as zero",
			input: []string{"1.1", "1", "1.0.1"},
			want:  []string{"1", "1.0.1", "1.1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := make([]semVer, 0, len(test.input))
			for _, v := range test.input {
				input = append(input, newSemVer(v))
			}

			got := sortSemVer(input)

			if len(got) != len(test.want) {
				t.Fatalf("sortSemVer(%q) returned %d items, want %d", test.input, len(got), len(test.want))
			}

			for i, v := range got {
				if v.original != test.want[i] {
					t.Errorf("sortSemVer(%q)[%d] = %s, want %s", test.input, i, v.original, test.want[i])
				}
			}
		})
	}
}
