package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/go-chi/chi/v5"
)

func (app *Application) GetBoard() http.HandlerFunc {
	requiredTemplates := []string{"board"}

	tmpl, err := app.createTemplate(requiredTemplates)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		boardId := chi.URLParam(r, "boardId")
		threads, err := app.ThreadModel.GetLatest(boardId)
		if err != nil {
			fmt.Printf("Error getting threads: %s\n", err.Error())
		}

		threadsTemplate := []struct {
			Thread  models.Thread
			Replies []models.Reply
		}{}

		for _, v := range threads {
			replies, _ := app.ReplyModel.GetLatestReplies(v.BoardID, int(v.ID), 5)
			threadsTemplate = append(threadsTemplate, struct {
				Thread  models.Thread
				Replies []models.Reply
			}{
				Thread:  v,
				Replies: replies,
			})
		}

		templateData := map[string]interface{}{
			"Threads": threadsTemplate,
		}
		tmpl.ExecuteTemplate(w, "base", &templateData)
	}
}
