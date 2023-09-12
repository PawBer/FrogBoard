package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/dchest/captcha"
	"github.com/go-chi/chi/v5"
)

func (app *Application) GetBoard(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"board"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

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

	boards := templateData["Boards"].([]models.Board)
	var board models.Board

	for _, v := range boards {
		if v.ID == boardId {
			board = v
		}
	}

	templateData["Board"] = board
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

func (app *Application) GetBoardEdit(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"boardedit"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	boardId := chi.URLParam(r, "boardId")

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if app.Sessions.Exists(r.Context(), "form-id") && app.Sessions.Exists(r.Context(), "form-fullname") && app.Sessions.Exists(r.Context(), "form-bumplimit") {
		templateData["FormID"] = app.Sessions.PopString(r.Context(), "form-id")
		templateData["FormFullName"] = app.Sessions.PopString(r.Context(), "form-fullname")
		templateData["FormBumpLimit"] = app.Sessions.PopString(r.Context(), "form-bumplimit")
	}

	boards := templateData["Boards"].([]models.Board)
	var board models.Board

	for _, v := range boards {
		if v.ID == boardId {
			board = v
		}
	}

	templateData["Board"] = board

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostBoardEdit(w http.ResponseWriter, r *http.Request) {
	formModel := struct {
		ID        string `form:"board-id"`
		FullName  string `form:"full-name"`
		BumpLimit string `form:"bump-limit"`
	}{}

	r.ParseForm()
	err := app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	bumpLimit, err := strconv.ParseInt(formModel.BumpLimit, 10, 32)
	if err != nil {
		app.serverError(w, err)
		return
	}

	newBoard := models.Board{
		ID:        formModel.ID,
		FullName:  formModel.FullName,
		BumpLimit: uint(bumpLimit),
	}

	err = app.BoardModel.Update(newBoard)
	if err != nil {
		app.Sessions.Put(r.Context(), "flash", "Something went wrong while editing the board")

		app.Sessions.Put(r.Context(), "form-id", formModel.ID)
		app.Sessions.Put(r.Context(), "form-fullname", formModel.FullName)
		app.Sessions.Put(r.Context(), "form-bumplimit", formModel.BumpLimit)

		url := fmt.Sprintf("/admin/board/%s/edit/", formModel.ID)
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "Board edited succesfully")
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (app *Application) GetBoardDelete(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"boarddelete"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	boardId := chi.URLParam(r, "boardId")

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	boards := templateData["Boards"].([]models.Board)
	var board models.Board

	for _, v := range boards {
		if v.ID == boardId {
			board = v
		}
	}

	templateData["Board"] = board

	tmpl = tmpl.Funcs(app.getFuncs(r))

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostBoardDelete(w http.ResponseWriter, r *http.Request) {
	boardId := chi.URLParam(r, "boardId")

	err := app.BoardModel.Delete(boardId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.FileInfoModel.DeleteOrphanedFiles()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "Board deleted succesfully")
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (app *Application) GetBoardCreate(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"boardcreate"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if app.Sessions.Exists(r.Context(), "form-id") && app.Sessions.Exists(r.Context(), "form-fullname") && app.Sessions.Exists(r.Context(), "form-bumplimit") {
		templateData["FormID"] = app.Sessions.PopString(r.Context(), "form-id")
		templateData["FormFullName"] = app.Sessions.PopString(r.Context(), "form-fullname")
		templateData["FormBumpLimit"] = app.Sessions.PopString(r.Context(), "form-bumplimit")
	}

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostBoardCreate(w http.ResponseWriter, r *http.Request) {
	formModel := struct {
		ID        string `form:"board-id"`
		FullName  string `form:"full-name"`
		BumpLimit string `form:"bump-limit"`
	}{}

	r.ParseForm()
	err := app.FormDecoder.Decode(&formModel, r.Form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	bumpLimit, err := strconv.ParseInt(formModel.BumpLimit, 10, 32)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.BoardModel.Insert(formModel.ID, formModel.FullName, uint(bumpLimit))
	if err != nil {
		app.Sessions.Put(r.Context(), "flash", "Something went wrong while editing the board")

		app.Sessions.Put(r.Context(), "form-id", formModel.ID)
		app.Sessions.Put(r.Context(), "form-fullname", formModel.FullName)
		app.Sessions.Put(r.Context(), "form-bumplimit", formModel.BumpLimit)

		http.Redirect(w, r, "/admin/board/create/", http.StatusSeeOther)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "Board edited succesfully")
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}
