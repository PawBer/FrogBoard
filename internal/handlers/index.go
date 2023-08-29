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
		boards, err := app.BoardModel.GetBoards()
		if err != nil {
			app.serverError(w, err)
			return
		}
		data := map[string]interface{}{
			"Boards": boards,
		}

		err = tmpl.ExecuteTemplate(w, "base", &data)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
}
