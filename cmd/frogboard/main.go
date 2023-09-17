package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/PawBer/FrogBoard/internal/handlers"
	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/PawBer/FrogBoard/pkg/filestorage"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/go-chi/chi/v5"
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

type Config struct {
	Port        string
	Db          DbConfig
	Redis       RedisConfig
	FileStorage FileStorage
}

type DbConfig struct {
	Hostname  string
	Port      string
	TableName string
	Username  string
	Password  string
}

type RedisConfig struct {
	Hostname string
	Port     string
}

type FileStorage struct {
	Type string
	Fs   struct {
		Path string
	}
}

func main() {
	infoLog := log.New(os.Stdout, "INFO ", log.Ltime)
	errorLog := log.New(os.Stderr, "WARNING ", log.Ldate|log.Ltime|log.Lshortfile)

	if _, err := os.Stat("/var/frogboard/config.toml"); err != nil {
		router := chi.NewRouter()
		router.Get("/", GetFirstRun(templates))
		router.Post("/", PostFirstRun)

		var port string
		if os.Getenv("FROGBOARD_PORT") != "" {
			port = os.Getenv("FROGBOARD_PORT")
		} else {
			port = "7543"
		}

		log.Printf("Please go to localhost:%s to start initial setup", port)

		listenAddress := fmt.Sprintf(":%s", port)
		log.Fatal(http.ListenAndServe(listenAddress, router))
		return
	}

	config := Config{}
	_, err := toml.DecodeFile("/var/frogboard/config.toml", &config)
	if err != nil {
		log.Fatalf("Error parsing config: %s", err.Error())
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", config.Db.Hostname, config.Db.Port, config.Db.Username, config.Db.TableName, config.Db.Password)
	dbConn, _ := sql.Open("postgres", connStr)
	err = dbConn.Ping()
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

	var fileStore filestorage.FileStore
	if config.FileStorage.Type == "fs" {
		fileStore = filestorage.NewFileSystemStore(config.FileStorage.Fs.Path)
	}

	redisConnStr := fmt.Sprintf("%s:%s", config.Redis.Hostname, config.Redis.Port)
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisConnStr)
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

	var port string
	if os.Getenv("FROGBOARD_PORT") != "" {
		port = os.Getenv("FROGBOARD_PORT")
	} else {
		port = config.Port
	}

	log.Printf("Starting server at :%s", port)

	listenAddress := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(listenAddress, app.GetRouter()))
}
