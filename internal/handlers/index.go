package handlers

import (
	"log"
	"net/http"
)

func (app *Application) GetIndex(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"index"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	tmpl = tmpl.Funcs(app.getFuncs(r))

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}
