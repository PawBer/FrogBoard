package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

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
				http.NotFound(w, r)
				return
			}

			url := fmt.Sprintf("/%s/%d/#p%d", reply.BoardID, reply.ThreadID, reply.ID)
			http.Redirect(w, r, url, http.StatusPermanentRedirect)
			return
		}

		files, err := app.FileInfoModel.GetFilesForPost(boardId, thread.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(app.ErrorLog.Writer(), "Error getting files for thread: %s\n", err.Error())
			fmt.Fprint(w, "Could not get files for thread")
			return
		}
		thread.Files = files

		replies, err := app.ReplyModel.GetRepliesToPost(boardId, uint(postId))
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(app.ErrorLog.Writer(), "Error getting replies to thread: %s\n", err.Error())
			fmt.Fprint(w, "Could not get replies")
			return
		}

		for i := 0; i < len(replies); i++ {
			files, err := app.FileInfoModel.GetFilesForPost(boardId, replies[i].ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(app.ErrorLog.Writer(), "Error getting files for reply: %s\n", err.Error())
				fmt.Fprint(w, "Could not get files for reply")
				return
			}
			replies[i].Files = files
		}

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

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(app.ErrorLog.Writer(), "Error parsing form: %s\n", err.Error())
		fmt.Fprint(w, "Form error")
		return
	}

	err = app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(app.ErrorLog.Writer(), "Error parsing form: %s\n", err.Error())
		fmt.Fprint(w, "Form error")
		return
	}

	var fileKeys []string

	files := r.MultipartForm.File["files"]

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(app.ErrorLog.Writer(), "Error opening form files: %s\n", err.Error())
			fmt.Fprint(w, "Bad files in form")
			return
		}
		defer file.Close()

		buf, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(app.ErrorLog.Writer(), "Error reading form files: %s\n", err.Error())
			fmt.Fprint(w, "Bad files in form")
			return
		}

		key, err := app.FileInfoModel.InsertFile(fileHeader.Filename, buf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(app.ErrorLog.Writer(), "Error uploading file from form: %s\n", err.Error())
			fmt.Fprint(w, "Bad files in form")
			return
		}

		fileKeys = append(fileKeys, key)
	}

	postId, err := app.ReplyModel.Insert(boardId, uint(threadId), formModel.Content, fileKeys)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(app.ErrorLog.Writer(), "Error inserting reply: %s\n", err.Error())
		fmt.Fprint(w, "Insert error")
		return
	}

	url := fmt.Sprintf("/%s/%d/", boardId, postId)
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
