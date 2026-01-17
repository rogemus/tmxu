package cli

import (
	"sort"
	"strconv"
	"strings"
)

type semVer struct {
	original string
	major    int
	minor    int
	patch    int
}

func newSemVer(v string) semVer {
	sv := semVer{original: strings.TrimSpace(v)}
	parts := strings.Split(v, ".")
	v = strings.TrimPrefix(v, "v")

	if len(parts) > 0 {
		sv.major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 {
		sv.minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		sv.patch, _ = strconv.Atoi(parts[2])
	}

	return sv
}

func sortSemVer(semVers []semVer) []semVer {
	sort.Slice(semVers, func(i, j int) bool {
		if semVers[i].major != semVers[j].major {
			return semVers[i].major < semVers[j].major
		}

		if semVers[i].minor != semVers[j].minor {
			return semVers[i].minor < semVers[j].minor
		}

		return semVers[i].patch < semVers[j].patch
	})

	return semVers
}
