package handlers

import (
	"fmt"
	"html/template"
	"io/fs"
)

func (app *Application) createTemplate(requiredTemplates []string) (*template.Template, error) {
	var templateFileNames []string

	requiredFiles, err := fs.ReadDir(app.Templates, "templates/required")
	if err != nil {
		return nil, err
	}

	for _, template := range requiredFiles {
		templateFileNames = append(templateFileNames, fmt.Sprintf("templates/required/%s", template.Name()))
	}

	for _, templateName := range requiredTemplates {
		filePath := fmt.Sprintf("templates/%s.tmpl.html", templateName)
		templateFileNames = append(templateFileNames, filePath)
	}

	tmpl, err := template.ParseFS(app.Templates, templateFileNames...)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
