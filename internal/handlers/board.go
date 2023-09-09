package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/dchest/captcha"
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

		captchaId := captcha.New()

		templateData, err := app.getTemplateData(r)
		if err != nil {
			app.serverError(w, err)
			return
		}

		templateData["BoardID"] = boardId
		templateData["Threads"] = threads
		templateData["CaptchaID"] = captchaId

		if app.Sessions.Exists(r.Context(), "form-title") && app.Sessions.Exists(r.Context(), "form-content") {
			templateData["FormTitle"] = app.Sessions.PopString(r.Context(), "form-title")
			templateData["FormContent"] = app.Sessions.PopString(r.Context(), "form-content")
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
		Title       string `form:"title"`
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

		app.Sessions.Put(r.Context(), "form-title", formModel.Title)
		app.Sessions.Put(r.Context(), "form-content", formModel.Content)

		url := fmt.Sprintf("/%s/", boardId)
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

	postId, err := app.ThreadModel.Insert(boardId, formModel.Title, formModel.Content, fileInfos)
	if err != nil {
		app.serverError(w, err)
		return
	}

	url := fmt.Sprintf("/%s/%d/#p%d", boardId, postId, postId)
	http.Redirect(w, r, url, http.StatusFound)
}
