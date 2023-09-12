package handlers

import (
	"embed"
	"log"
	"net/http"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/PawBer/FrogBoard/pkg/filestorage"
	"github.com/alexedwards/scs/v2"
	"github.com/dchest/captcha"
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
	CitationModel *models.CitationModel
	UserModel     *models.UserModel
	Templates     embed.FS
	Public        embed.FS
	FormDecoder   *form.Decoder
	FileStore     filestorage.FileStore
	Sessions      *scs.SessionManager
}

func (app *Application) GetRouter() http.Handler {
	router := chi.NewRouter()

	// Middleware
	router.Use(app.Logging)
	router.Use(app.Sessions.LoadAndSave)

	router.Get("/public/*", app.GetPublic())

	router.Get("/", app.GetIndex)
	router.Get("/login/", app.GetLogin)
	router.Post("/login/", app.PostLogin)
	router.Post("/logout/", app.PostLogout)
	router.Get("/{boardId}/", app.GetBoard)
	router.Post("/{boardId}/", app.PostBoard)
	router.Get("/{boardId}/{postId}/", app.GetPost)
	router.Post("/{boardId}/{postId}/", app.PostThread)
	router.Get("/file/{hash}/", app.GetFile)
	router.Get("/file/{hash}/thumb/", app.GetFileThumbnail)
	router.Mount("/captcha/", captcha.Server(240, 80))
	router.Get("/api/post/{boardId}/{postId}/", app.GetPostJson)

	router.Mount("/admin/", app.getAdminRouter())

	return router
}

func (app *Application) getAdminRouter() http.Handler {
	router := chi.NewRouter()

	router.Use(app.AdminOnly)

	router.Get("/", app.GetAdmin)
	router.Get("/board/create/", app.GetBoardCreate)
	router.Post("/board/create/", app.PostBoardCreate)
	router.Get("/board/{boardId}/edit/", app.GetBoardEdit)
	router.Post("/board/{boardId}/edit/", app.PostBoardEdit)
	router.Get("/board/{boardId}/delete/", app.GetBoardDelete)
	router.Post("/board/{boardId}/delete/", app.PostBoardDelete)
	router.Get("/{boardId}/{postId}/delete/", app.GetDelete)
	router.Post("/{boardId}/{postId}/delete/", app.PostDelete)

	return router
}
