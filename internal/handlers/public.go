package handlers

import (
	"io/fs"
	"log"
	"net/http"
)

func (app *Application) GetPublic() http.Handler {
	serverRoot, err := fs.Sub(app.Public, "public")
	if err != nil {
		log.Fatalf("Problem with public dir: %s", err.Error())
	}

	fileServer := http.FileServer(http.FS(serverRoot))
	return http.StripPrefix("/public", fileServer)
}
