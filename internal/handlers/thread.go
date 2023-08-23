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

func (app *Application) GetThread() http.HandlerFunc {
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

		templateData := map[string]interface{}{
			"Thread":  thread,
			"Replies": replies,
		}
		tmpl.ExecuteTemplate(w, "base", &templateData)
	}
}
