package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/PawBer/FrogBoard/internal/handlers"
	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/PawBer/FrogBoard/pkg/filestorage"
	"github.com/go-playground/form"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed templates
var templates embed.FS

//go:embed public
var public embed.FS

func main() {
	infoLog := log.New(os.Stdout, "INFO ", log.Ltime)
	errorLog := log.New(os.Stderr, "WARNING ", log.Ldate|log.Ltime|log.Lshortfile)

	connStr := "host=localhost user=frogboard dbname=frogboard password=frogboardpassword sslmode=disable"
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Error connecting to db: %s", err.Error())
	}

	formDecoder := form.NewDecoder()

	fileStore := filestorage.NewFileSystemStore("files")

	db.AutoMigrate(&models.Board{}, &models.Thread{}, &models.Reply{})

	app := handlers.Application{
		InfoLog:     infoLog,
		ErrorLog:    errorLog,
		BoardModel:  &models.BoardModel{DbConn: db},
		ThreadModel: &models.ThreadModel{DbConn: db},
		ReplyModel:  &models.ReplyModel{DbConn: db},
		Templates:   templates,
		Public:      public,
		FormDecoder: formDecoder,
		FileStore:   fileStore,
	}

	log.Printf("Starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", app.GetRouter()))
}
