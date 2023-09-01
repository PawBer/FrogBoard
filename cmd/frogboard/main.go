package main

import (
	"database/sql"
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/PawBer/FrogBoard/internal/handlers"
	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/PawBer/FrogBoard/pkg/filestorage"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/go-playground/form"
	_ "github.com/lib/pq"
)

//go:embed templates
var templates embed.FS

//go:embed public
var public embed.FS

func main() {
	infoLog := log.New(os.Stdout, "INFO ", log.Ltime)
	errorLog := log.New(os.Stderr, "WARNING ", log.Ldate|log.Ltime|log.Lshortfile)

	connStr := "host=localhost user=frogboard dbname=frogboard password=frogboardpassword sslmode=disable"
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to db: %s", err.Error())
	}

	db := goqu.Dialect("postgres").DB(dbConn)

	formDecoder := form.NewDecoder()

	fileStore := filestorage.NewFileSystemStore("files")

	boardModel := &models.BoardModel{DbConn: db}
	fileInfoModel := &models.FileInfoModel{DbConn: db, FileStore: fileStore}
	citationModel := &models.CitationModel{DbConn: db}

	replyModel := &models.ReplyModel{
		DbConn:        db,
		FileInfoModel: fileInfoModel,
		CitationModel: citationModel,
	}
	threadModel := &models.ThreadModel{
		DbConn:        db,
		FileInfoModel: fileInfoModel,
		CitationModel: citationModel,
		ReplyModel:    replyModel,
	}

	app := handlers.Application{
		InfoLog:       infoLog,
		ErrorLog:      errorLog,
		BoardModel:    boardModel,
		ThreadModel:   threadModel,
		ReplyModel:    replyModel,
		FileInfoModel: fileInfoModel,
		CitationModel: citationModel,
		Templates:     templates,
		Public:        public,
		FormDecoder:   formDecoder,
		FileStore:     fileStore,
	}

	log.Printf("Starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", app.GetRouter()))
}
