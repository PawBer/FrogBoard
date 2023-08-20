package handlers

import (
	"embed"
	"log"
	"net/http"

	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
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
	router := httprouter.New()

	router.Handler(http.MethodGet, "/public/*filepath", app.GetPublic())

	router.HandlerFunc(http.MethodGet, "/", app.GetIndex())

	// Setting up middleware
	logging := app.Logging

	return alice.New(logging).Then(router)
}
