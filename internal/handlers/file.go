package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Application) GetFile(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	file, err := app.FileStore.GetFile(hash)
	if err != nil {
		app.notFound(w)
		return
	}

	w.Write(file)
}

func (app *Application) GetFileThumbnail(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	file, err := app.FileStore.GetFileThumbnail(hash)
	if err != nil {
		app.notFound(w)
		return
	}

	w.Write(file)
}

func (app *Application) GetFileDelete(w http.ResponseWriter, r *http.Request) {
	requiredTemplates := []string{"filedelete"}

	tmpl, err := app.createTemplate(requiredTemplates, r)
	if err != nil {
		log.Fatalf("Failed to load templates: %s", err.Error())
	}

	fileId := chi.URLParam(r, "fileId")

	templateData, err := app.getTemplateData(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData["ID"] = fileId

	err = tmpl.ExecuteTemplate(w, "base", &templateData)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) PostFileDelete(w http.ResponseWriter, r *http.Request) {
	fileId := chi.URLParam(r, "fileId")

	err := app.FileInfoModel.Delete(fileId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.Sessions.Put(r.Context(), "flash", "File deleted succesfully")
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}
