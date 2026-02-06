package cli

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func getVersionsTags() ([]string, error) {
	res, err := http.Get("https://proxy.golang.org/github.com/rogemus/tmxu/@v/list")

	if err != nil {
		fmt.Printf("[%v]\n", res)
		return nil, fmt.Errorf("unable to fetch tags from Proxy \n")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	return strings.Split(
		strings.TrimSpace(string(body)),
		"\n",
	), nil
}

func getNewestVersion() (semVer, error) {
	tags, err := getVersionsTags()
	if err != nil {
		return semVer{}, err
	}

	var versions []semVer
	for _, tag := range tags {
		versions = append(versions, newSemVer(tag))
	}

	sorted := sortSemVer(versions)
	return sorted[len(sorted)-1], nil
}
