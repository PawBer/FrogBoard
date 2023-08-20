package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/PawBer/FrogBoard/internal/handlers"
	"github.com/PawBer/FrogBoard/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed templates
var templates embed.FS

//go:embed public
var public embed.FS

func main() {
	infoLog := log.New(os.Stdout, "INFO ", log.Ltime)
	errorLog := log.New(os.Stderr, "WARNING ", log.Ldate|log.Ltime|log.Lshortfile)

	connStr := "host=localhost user=frogboard dbname=frogboard password=frogboardpassword sslmode=disable"
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to db: %s", err.Error())
	}

	db.AutoMigrate(&models.Board{}, &models.Thread{}, &models.Reply{})

	app := handlers.Application{
		InfoLog:     infoLog,
		ErrorLog:    errorLog,
		BoardModel:  &models.BoardModel{DbConn: db},
		ThreadModel: &models.ThreadModel{DbConn: db},
		ReplyModel:  &models.ReplyModel{DbConn: db},
		Templates:   templates,
		Public:      public,
	}

	log.Printf("Starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", app.GetRouter()))
}
