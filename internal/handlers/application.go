package handlers

import (
	"embed"
	"log"
	"net/http"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/go-chi/chi/v5"
)

type Application struct {
	InfoLog     *log.Logger
	ErrorLog    *log.Logger
	BoardModel  *models.BoardModel
	ThreadModel *models.ThreadModel
	ReplyModel  *models.ReplyModel
	Templates   embed.FS
	Public      embed.FS
}

func (app *Application) GetRouter() http.Handler {
	router := chi.NewRouter()

	// Middleware
	router.Use(app.Logging)

	router.Get("/public/*", app.GetPublic())

	router.Get("/", app.GetIndex())
	router.Get("/{boardId}/", app.GetBoard())
	router.Mount("/{boardId}", app.GetBoard())

	return router
}
