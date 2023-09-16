package main

import (
	"database/sql"
	"embed"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/PawBer/FrogBoard/internal/handlers"
	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/PawBer/FrogBoard/pkg/filestorage"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/go-playground/form"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
)

//go:embed templates
var templates embed.FS

//go:embed public
var public embed.FS

//go:embed migrations
var migrations embed.FS

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	infoLog := log.New(os.Stdout, "INFO ", log.Ltime)
	errorLog := log.New(os.Stderr, "WARNING ", log.Ldate|log.Ltime|log.Lshortfile)

	connStr := "host=db user=frogboard dbname=frogboard password=frogboardpassword sslmode=disable"
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to db: %s", err.Error())
	}

	driver, _ := postgres.WithInstance(dbConn, &postgres.Config{})
	source, _ := iofs.New(migrations, "migrations")

	migrator, _ := migrate.NewWithInstance("iofs", source, "postgres", driver)
	infoLog.Output(2, "Starting migration")
	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Error migrating: %s", err.Error())
	}

	db := goqu.Dialect("postgres").DB(dbConn)

	formDecoder := form.NewDecoder()

	fileStore := filestorage.NewFileSystemStore("/var/frogboard/filestorage")

	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis:6379")
		},
	}

	sessionStore := scs.New()
	sessionStore.Lifetime = 24 * 7 * time.Hour
	sessionStore.Store = redisstore.New(pool)

	boardModel := &models.BoardModel{DbConn: db}
	fileInfoModel := &models.FileInfoModel{DbConn: db, FileStore: fileStore}
	citationModel := &models.CitationModel{DbConn: db}

	userModel := &models.UserModel{
		DbConn: db,
	}

	banModel := &models.BanModel{
		DbConn: db,
	}

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
		UserModel:     userModel,
		BanModel:      banModel,
		Templates:     templates,
		Public:        public,
		FormDecoder:   formDecoder,
		FileStore:     fileStore,
		Sessions:      sessionStore,
	}

	log.Printf("Starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", app.GetRouter()))
}
