package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

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

		boards, err := app.BoardModel.GetBoards()
		if err != nil {
			app.serverError(w, err)
			return
		}

		templateData := map[string]interface{}{
			"BoardID": boardId,
			"Boards":  boards,
			"Threads": threads,
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
