package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/go-chi/chi/v5"
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
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			reply, err := app.ReplyModel.Get(boardId, uint(postId))
			if err != nil && errors.Is(err, sql.ErrNoRows) {
				app.notFound(w)
				return
			}

			url := fmt.Sprintf("/%s/%d/#p%d", reply.BoardID, reply.ThreadID, reply.ID)
			http.Redirect(w, r, url, http.StatusPermanentRedirect)
			return
		}
		if err != nil {
			app.serverError(w, err)
			return
		}

		if err = app.populateFieldsInPost(&thread.Post); err != nil {
			app.serverError(w, err)
			return
		}

		replies, err := app.ReplyModel.GetRepliesToPost(boardId, uint(postId))
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			app.serverError(w, err)
			return
		}

		for i := 0; i < len(replies); i++ {
			if err = app.populateFieldsInPost(&replies[i].Post); err != nil {
				app.serverError(w, err)
				return
			}
		}

		boards, err := app.BoardModel.GetBoards()
		if err != nil {
			app.serverError(w, err)
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
			app.serverError(w, err)
			return
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

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var fileKeys []string

	files := r.MultipartForm.File["files"]

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer file.Close()

		buf, err := io.ReadAll(file)
		if err != nil {
			app.serverError(w, err)
			return
		}

		key, err := app.FileInfoModel.InsertFile(fileHeader.Filename, buf)
		if err != nil {
			app.serverError(w, err)
			return
		}

		fileKeys = append(fileKeys, key)
	}

	postId, err := app.ReplyModel.Insert(boardId, uint(threadId), formModel.Content, fileKeys)
	if err != nil {
		app.serverError(w, err)
		return
	}

	citations := models.GetCitations(boardId, postId, formModel.Content)
	for _, citation := range citations {
		if err := app.CitationModel.InsertCitation(citation.BoardID, citation.PostID, citation.Cites); err != nil {
			app.serverError(w, err)
			return
		}
	}

	url := fmt.Sprintf("/%s/%d/", boardId, postId)
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
