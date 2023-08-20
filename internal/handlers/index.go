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
		threads, _ := app.ThreadModel.GetLatest("b")
		data := map[string]interface{}{
			"Threads": threads,
		}
		tmpl.ExecuteTemplate(w, "base", &data)
	}
}
