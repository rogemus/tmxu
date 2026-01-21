package cli

import "fmt"

type tTemplate = tSession

func listTemplates() ([]tTemplate, error) {
	ts, err := loadTemplateFiles()
	if err != nil {
		return nil, fmt.Errorf("Cannot load template files from `~/.config/tmxu/templates` \n")
	}

	return ts, nil
}
