package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (app *Application) GetPostJson(w http.ResponseWriter, r *http.Request) {
	boardId := chi.URLParam(r, "boardId")
	postIdStr := chi.URLParam(r, "postId")
	postId, _ := strconv.ParseUint(postIdStr, 10, 32)

	thread, err := app.ThreadModel.Get(boardId, uint(postId))
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		reply, err := app.ReplyModel.Get(boardId, uint(postId))
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		reply.Content = string(reply.FormatedContent())

		json.NewEncoder(w).Encode(&reply)
		return
	}
	json.NewEncoder(w).Encode(&thread)
}

func (app *Application) DeletePost(w http.ResponseWriter, r *http.Request) {
	boardId := chi.URLParam(r, "boardId")
	postIdStr := chi.URLParam(r, "postId")
	postId, _ := strconv.ParseUint(postIdStr, 10, 32)

	if err := app.ThreadModel.Delete(boardId, uint(postId)); err != nil {
		app.serverError(w, err)
		return
	}

	if _, err := app.ReplyModel.Delete(boardId, uint(postId)); err != nil {
		app.serverError(w, err)
		return
	}
}
