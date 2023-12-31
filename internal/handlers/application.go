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
	BanModel      *models.BanModel
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
	router.Use(app.BlockBannedUsers)
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
	router.Get("/file/{fileId}/delete/", app.GetFileDelete)
	router.Post("/file/{fileId}/delete/", app.PostFileDelete)
	router.Get("/bans/", app.GetBans)
	router.Get("/bans/create/", app.GetBanCreate)
	router.Post("/bans/create/", app.PostBanCreate)
	router.Get("/bans/{ip}/delete/", app.GetBanDelete)
	router.Post("/bans/{ip}/delete/", app.PostBanDelete)
	router.Get("/users/create/", app.GetUserCreate)
	router.Post("/users/create/", app.PostUserCreate)
	router.Get("/users/create/success/", app.GetUserCreateSuccess)
	router.Get("/users/{username}/edit/", app.GetUserEdit)
	router.Post("/users/{username}/edit/", app.PostUserEdit)
	router.Get("/users/{username}/delete/", app.GetUserDelete)
	router.Post("/users/{username}/delete/", app.PostUserDelete)
	router.Get("/users/{username}/passwordreset/", app.GetPasswordReset)
	router.Post("/users/{username}/passwordreset/", app.PostPasswordReset)
	router.Get("/users/{username}/passwordreset/success/", app.GetPasswordResetSuccess)
	router.Get("/users/passwordchange/", app.GetPasswordChange)
	router.Post("/users/passwordchange/", app.PostPasswordChange)

	return router
}
