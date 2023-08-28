package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Application) GetFile(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	file, err := app.FileStore.GetFile(hash)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Write(file)
}

func (app *Application) GetFileThumbnail(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	file, err := app.FileStore.GetFileThumbnail(hash)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Write(file)
}
