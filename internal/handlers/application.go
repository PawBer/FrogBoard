package handlers

import (
	"log"

	"github.com/julienschmidt/httprouter"
)

type Application struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func (app *Application) GetRouter() *httprouter.Router {
	router := httprouter.New()
	router.GET("/", app.Index)

	return router
}
