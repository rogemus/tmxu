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

func saveTemplate(session tSession, path string) error {
	t, err := sessionToTemplate(session, path)
	if err != nil {
		return fmt.Errorf("Cannot convert session: %s to template \n", t.Name)
	}

	err = saveTemplateFile(t)
	if err != nil {
		return fmt.Errorf("Cannot save template to file at `~/.config/tmxu/templates/%s.json` \n", t.Name)
	}

	return nil
}

func deleteTemplate(templateName string) error {
	err := deleteTemplateFile(templateName)
	if err != nil {
		return fmt.Errorf("Cannot delete template file: `~/.config/tmxu/templates/%s.json` \n", templateName)
	}

	return nil
}

func sessionToTemplate(session tSession, path string) (tTemplate, error) {
	for i, w := range session.Windows {
		for j, p := range w.Panes {
			p.Path = "PATH"
			w.Panes[j] = p
		}
		session.Windows[i] = w
	}

	return tTemplate(session), nil
}
