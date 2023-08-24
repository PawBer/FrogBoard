package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func (app *Application) GetPostJson(w http.ResponseWriter, r *http.Request) {
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

		json.NewEncoder(w).Encode(&reply)
		return
	}
	json.NewEncoder(w).Encode(&thread)
}
