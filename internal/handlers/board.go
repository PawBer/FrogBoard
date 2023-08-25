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
			"BoardID": boardId,
			"Threads": threadsTemplate,
		}
		err = tmpl.ExecuteTemplate(w, "base", &templateData)
		if err != nil {
			app.ErrorLog.Printf("Error executing template: %s\n", err.Error())
		}
	}
}

func (app *Application) PostBoard(w http.ResponseWriter, r *http.Request) {
	boardId := chi.URLParam(r, "boardId")

	formModel := struct {
		Title   string `form:"title"`
		Content string `form:"content"`
	}{}

	r.ParseForm()

	err := app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(app.ErrorLog.Writer(), "Error parsing form: %s\n", err.Error())
		fmt.Fprint(w, "Form error")
		return
	}

	postId, err := app.ThreadModel.Insert(boardId, formModel.Title, formModel.Content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(app.ErrorLog.Writer(), "Error inserting thread: %s\n", err.Error())
		fmt.Fprint(w, "Insert error")
		return
	}

	url := fmt.Sprintf("/%s/%d/", boardId, postId)
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
