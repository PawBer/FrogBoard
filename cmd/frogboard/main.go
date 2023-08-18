package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/PawBer/FrogBoard/internal/handlers"
)

//go:embed templates
var templates embed.FS

//go:embed public
var public embed.FS

func main() {
	infoLog := log.New(os.Stdout, "INFO ", log.Ltime)
	errorLog := log.New(os.Stderr, "WARNING ", log.Ldate|log.Ltime|log.Lshortfile)

	app := handlers.Application{
		InfoLog:   infoLog,
		ErrorLog:  errorLog,
		Templates: templates,
		Public:    public,
	}

	log.Printf("Starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", app.GetRouter()))
}
