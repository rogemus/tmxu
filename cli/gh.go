package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ghTag struct {
	Name string `json:"name"`
}

func getGHtags() ([]ghTag, error) {
	res, err := http.Get("https://api.github.com/repos/rogemus/tmxu/tags")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch tags from Github \n")
	}
	defer res.Body.Close()

	var tags []ghTag

	err = json.NewDecoder(res.Body).Decode(&tags)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal sassion data")
	}

	return tags, nil
}

func getNewestVersion() (semVer, error) {
	tags, err := getGHtags()
	if err != nil {
		return semVer{}, err
	}

	var versions []semVer
	for _, tag := range tags {
		versions = append(versions, newSemVer(tag.Name))
	}

	sorted := sortSemVer(versions)
	return sorted[len(sorted)-1], nil
}
