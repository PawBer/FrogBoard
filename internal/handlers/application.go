package handlers

import (
	"embed"
	"log"
	"net/http"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/PawBer/FrogBoard/pkg/filestorage"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form"
)

type Application struct {
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	BoardModel    *models.BoardModel
	ThreadModel   *models.ThreadModel
	ReplyModel    *models.ReplyModel
	FileInfoModel *models.FileInfoModel
	Templates     embed.FS
	Public        embed.FS
	FormDecoder   *form.Decoder
	FileStore     filestorage.FileStore
}

func (app *Application) GetRouter() http.Handler {
	router := chi.NewRouter()

	// Middleware
	router.Use(app.Logging)

	router.Get("/public/*", app.GetPublic())

	router.Get("/", app.GetIndex())
	router.Get("/{boardId}/", app.GetBoard())
	router.Post("/{boardId}/", app.PostBoard)
	router.Get("/{boardId}/{postId}/", app.GetPost())
	router.Post("/{boardId}/{postId}/", app.PostThread)
	router.Get("/file/{hash}/", app.GetFile)
	router.Get("/file/{hash}/thumb/", app.GetFileThumbnail)
	router.Get("/api/post/{boardId}/{postId}/", app.GetPostJson)

	return router
}
