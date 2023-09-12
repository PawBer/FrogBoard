package handlers

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"runtime/debug"

	"github.com/PawBer/FrogBoard/internal/models"
)

func (app *Application) getTemplateData(r *http.Request) (map[string]interface{}, error) {
	boards, err := app.BoardModel.GetBoards()
	if err != nil {
		return nil, err
	}

	flash := app.Sessions.PopString(r.Context(), "flash")

	templateData := map[string]interface{}{
		"Flash":  flash,
		"Boards": boards,
	}

	return templateData, nil
}

func (app *Application) createTemplate(requiredTemplates []string, r *http.Request) (*template.Template, error) {
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

	tmpl, err := template.New("").Funcs(app.getFuncs(r)).ParseFS(app.Templates, templateFileNames...)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func (app *Application) getFuncs(r *http.Request) template.FuncMap {
	return map[string]interface{}{
		"IsAuthenticated": func() bool {
			return app.Sessions.Exists(r.Context(), "authenticated")
		},
		"GetPermission": func() int {
			return app.Sessions.Get(r.Context(), "permission").(int)
		},
	}
}

func (app *Application) hasPermission(r *http.Request, perm models.UserPermission) bool {
	permission := models.UserPermission(app.Sessions.Get(r.Context(), "permission").(int))

	return permission == perm
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
