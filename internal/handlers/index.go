package handlers

import (
	"fmt"
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
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(app.ErrorLog.Writer(), "Could not get boards: %s\n", err.Error())
			fmt.Fprint(w, "Could not get boards")
			return
		}
		data := map[string]interface{}{
			"Boards": boards,
		}

		tmpl.ExecuteTemplate(w, "base", &data)
	}
}
