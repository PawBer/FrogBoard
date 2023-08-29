package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
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
			app.serverError(w, err)
			return
		}

		threadsTemplate := []struct {
			Thread  models.Thread
			Replies []models.Reply
		}{}

		for i := 0; i < len(threads); i++ {
			files, err := app.FileInfoModel.GetFilesForPost(boardId, threads[i].ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				app.serverError(w, err)
				return
			}
			threads[i].Files = files

			replies, err := app.ReplyModel.GetLatestReplies(threads[i].BoardID, int(threads[i].ID), 5)
			if err != nil {
				app.serverError(w, err)
				return
			}

			for j := 0; j < len(replies); j++ {
				files, err := app.FileInfoModel.GetFilesForPost(boardId, replies[j].ID)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					app.serverError(w, err)
					return
				}
				replies[j].Files = files
			}

			threadsTemplate = append(threadsTemplate, struct {
				Thread  models.Thread
				Replies []models.Reply
			}{
				Thread:  threads[i],
				Replies: replies,
			})
		}

		boards, err := app.BoardModel.GetBoards()
		if err != nil {
			app.serverError(w, err)
			return
		}

		templateData := map[string]interface{}{
			"BoardID": boardId,
			"Boards":  boards,
			"Threads": threadsTemplate,
		}
		err = tmpl.ExecuteTemplate(w, "base", &templateData)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
}

func (app *Application) PostBoard(w http.ResponseWriter, r *http.Request) {
	boardId := chi.URLParam(r, "boardId")

	formModel := struct {
		Title   string `form:"title"`
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

	postId, err := app.ThreadModel.Insert(boardId, formModel.Title, formModel.Content, fileKeys)
	if err != nil {
		app.serverError(w, err)
		return
	}

	url := fmt.Sprintf("/%s/%d/", boardId, postId)
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
