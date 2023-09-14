package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/dchest/captcha"
	"github.com/go-chi/chi/v5"
)

func (app *Application) GetPost(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"thread"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

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
		http.Redirect(w, r, url, http.StatusFound)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	captchaId := captcha.New()

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData["BoardID"] = boardId
	templateData["Thread"] = thread
	templateData["CaptchaID"] = captchaId

	if app.Sessions.Exists(r.Context(), "form-content") {
		templateData["FormContent"] = app.Sessions.PopString(r.Context(), "form-content")
	}

	tmpl = tmpl.Funcs(app.getFuncs(r))

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostThread(w http.ResponseWriter, r *http.Request) {
	boardId := chi.URLParam(r, "boardId")
	threadIdStr := chi.URLParam(r, "postId")
	threadId, _ := strconv.ParseUint(threadIdStr, 10, 32)

	formModel := struct {
		Content     string `form:"content"`
		CaptchaId   string `form:"captcha-id"`
		CaptchaCode string `form:"captcha-code"`
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

	if !captcha.VerifyString(formModel.CaptchaId, formModel.CaptchaCode) {
		app.Sessions.Put(r.Context(), "flash", "Failed captcha authentication")

		app.Sessions.Put(r.Context(), "form-content", formModel.Content)

		url := fmt.Sprintf("/%s/%d/", boardId, threadId)
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}

	var fileInfos []models.FileInfo

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

		fileInfo, err := app.FileInfoModel.InsertFile(fileHeader.Filename, buf)
		if err != nil {
			app.serverError(w, err)
			return
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	postId, err := app.ReplyModel.Insert(boardId, uint(threadId), formModel.Content, fileInfos, host)
	if err != nil {
		app.serverError(w, err)
		return
	}

	url := fmt.Sprintf("/%s/%d/#p%d", boardId, postId, postId)
	http.Redirect(w, r, url, http.StatusFound)
}
