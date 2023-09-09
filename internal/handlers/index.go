package handlers

import (
	"log"
	"net/http"
)

func (app *Application) GetIndex() http.HandlerFunc {
	requiredTemplates := []string{"index"}

	tmpl, err := app.createTemplate(requiredTemplates)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
}
