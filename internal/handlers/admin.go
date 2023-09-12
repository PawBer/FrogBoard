package handlers

import (
	"log"
	"net/http"
)

func (app *Application) GetAdmin(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"admin"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}