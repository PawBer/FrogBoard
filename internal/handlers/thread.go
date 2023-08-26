package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func (app *Application) GetPost() http.HandlerFunc {
	requiredTemplates := []string{"thread"}

	tmpl, err := app.createTemplate(requiredTemplates)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		boardId := chi.URLParam(r, "boardId")
		postIdStr := chi.URLParam(r, "postId")
		postId, _ := strconv.ParseUint(postIdStr, 10, 32)

		thread, err := app.ThreadModel.Get(boardId, uint(postId))
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			reply, err := app.ReplyModel.Get(boardId, uint(postId))
			if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
				http.NotFound(w, r)
				return
			}

			url := fmt.Sprintf("/%s/%d/#p%d", reply.BoardID, reply.ThreadID, reply.ID)
			http.Redirect(w, r, url, http.StatusPermanentRedirect)
			return
		}
		replies, _ := app.ReplyModel.GetRepliesToPost(boardId, uint(postId))

		boards, err := app.BoardModel.GetBoards()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(app.ErrorLog.Writer(), "Error getting boards: %s\n", err.Error())
			fmt.Fprint(w, "Could not get boards")
			return
		}

		templateData := map[string]interface{}{
			"BoardID": boardId,
			"Boards":  boards,
			"Thread":  thread,
			"Replies": replies,
		}

		err = tmpl.ExecuteTemplate(w, "base", &templateData)
		if err != nil {
			app.ErrorLog.Printf("Error executing template: %s\n", err.Error())
		}
	}
}

func (app *Application) PostThread(w http.ResponseWriter, r *http.Request) {
	boardId := chi.URLParam(r, "boardId")
	threadIdStr := chi.URLParam(r, "postId")
	threadId, _ := strconv.ParseUint(threadIdStr, 10, 32)

	formModel := struct {
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

	postId, err := app.ReplyModel.Insert(boardId, uint(threadId), formModel.Content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(app.ErrorLog.Writer(), "Error inserting reply: %s\n", err.Error())
		fmt.Fprint(w, "Insert error")
		return
	}

	url := fmt.Sprintf("/%s/%d/", boardId, postId)
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
