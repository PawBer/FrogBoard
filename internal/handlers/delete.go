package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (app *Application) GetDelete(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"delete"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	boardId := chi.URLParam(r, "boardId")
	postIdStr := chi.URLParam(r, "postId")
	postId, _ := strconv.ParseUint(postIdStr, 10, 32)

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	thread, err := app.ThreadModel.Get(boardId, uint(postId))
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		reply, err := app.ReplyModel.Get(boardId, uint(postId))
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			app.notFound(w)
			return
		}

		templateData["Post"] = reply
	} else {
		templateData["Post"] = thread
	}

	templateData["BoardID"] = boardId

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostDelete(w http.ResponseWriter, r *http.Request) {
	boardId := chi.URLParam(r, "boardId")
	postIdStr := chi.URLParam(r, "postId")
	postId, _ := strconv.ParseUint(postIdStr, 10, 32)

	threadId, err := app.ReplyModel.Delete(boardId, uint(postId))
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err := app.ThreadModel.Delete(boardId, uint(postId))
		if err != nil {
			app.serverError(w, err)
			return
		}

		err = app.FileInfoModel.DeleteOrphanedFiles()
		if err != nil {
			app.serverError(w, err)
			return
		}

		url := fmt.Sprintf("/%s/", boardId)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.FileInfoModel.DeleteOrphanedFiles()
	if err != nil {
		app.serverError(w, err)
		return
	}

	url := fmt.Sprintf("/%s/%d/", boardId, threadId)
	http.Redirect(w, r, url, http.StatusFound)
}
