package handlers

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"runtime/debug"
)

func (app *Application) getTemplateData() (map[string]interface{}, error) {
	boards, err := app.BoardModel.GetBoards()
	if err != nil {
		return nil, err
	}

	templateData := map[string]interface{}{
		"Boards": boards,
	}

	return templateData, nil
}

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

func (app *Application) serverError(w http.ResponseWriter, err error) {
	errorMessage := fmt.Sprintf("%s\n%s\n", err.Error(), debug.Stack())
	app.ErrorLog.Output(2, errorMessage)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *Application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
