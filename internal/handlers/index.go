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
		boards, _ := app.BoardModel.GetBoards()
		data := map[string]interface{}{
			"Boards": boards,
		}

		tmpl.ExecuteTemplate(w, "base", &data)
	}
}
