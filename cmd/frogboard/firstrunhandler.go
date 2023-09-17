package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/PawBer/FrogBoard/internal/models"
	"github.com/doug-martin/goqu/v9"
	"github.com/go-playground/form"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/gomodule/redigo/redis"
)

func GetFirstRun(templates embed.FS) http.HandlerFunc {
	html, err := templates.ReadFile("templates/firsttime.html")
	if err != nil {
		log.Fatalf("Error getting configuration page: %s\n", err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(html)
	}
}

func PostFirstRun(w http.ResponseWriter, r *http.Request) {
	formModel := struct {
		Port             string `form:"port"`
		PostgresHost     string `form:"postgres-host"`
		PostgresPort     string `form:"postgres-port"`
		PostgresDb       string `form:"postgres-db-name"`
		PostgresUsername string `form:"postgres-username"`
		PostgresPassword string `form:"postgres-password"`
		RedisHost        string `form:"redis-host"`
		RedisPort        string `form:"redis-port"`
		FileStoragePath  string `form:"fs-path"`
	}{}

	r.ParseForm()
	err := form.NewDecoder().Decode(&formModel, r.Form)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error decoding form: %s\n", err.Error())))
		return
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", formModel.PostgresHost, formModel.PostgresPort, formModel.PostgresUsername, formModel.PostgresDb, formModel.PostgresPassword)
	dbConn, _ := sql.Open("postgres", connStr)
	err = dbConn.Ping()
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Failed to connect to postgres: %s\n", err.Error())))
		return
	}

	w.Write([]byte("Postgres Connection successful\n"))

	driver, _ := postgres.WithInstance(dbConn, &postgres.Config{})
	source, _ := iofs.New(migrations, "migrations")

	migrator, _ := migrate.NewWithInstance("iofs", source, "postgres", driver)
	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		w.Write([]byte(fmt.Sprintf("Error migrating: %s\n", err.Error())))
	}

	userModel := models.UserModel{DbConn: goqu.Dialect("postgres").DB(dbConn)}
	password, err := userModel.RegisterUser("admin", "Administrator", models.Admin)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Failed to create admin user: %s\n", err.Error())))
		return
	}
	w.Write([]byte("Created admin user\n"))

	boardModel := models.BoardModel{DbConn: goqu.Dialect("postgres").DB(dbConn)}
	err = boardModel.Insert("b", "Random", 50)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Failed to create first board: %s\n", err.Error())))
		return
	}
	w.Write([]byte("/b/ Created\n"))

	dbConn.Close()

	redisConnStr := fmt.Sprintf("%s:%s", formModel.RedisHost, formModel.RedisPort)
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisConnStr)
		},
	}

	conn, err := pool.Dial()
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Connection to redis failed: %s\n", err.Error())))
		return
	}
	_, err = conn.Do("ping")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Connection to redis failed: %s\n", err.Error())))
		return
	}
	conn.Close()

	w.Write([]byte("Redis Connection successful\n"))

	w.Write([]byte("Everything is correct. Writing configuration\n"))

	config := Config{
		Port: formModel.Port,
		Db: DbConfig{
			Hostname:  formModel.PostgresHost,
			Port:      formModel.PostgresPort,
			TableName: formModel.PostgresDb,
			Username:  formModel.PostgresUsername,
			Password:  formModel.PostgresPassword,
		},
		Redis: RedisConfig{
			Hostname: formModel.RedisHost,
			Port:     formModel.RedisPort,
		},
		FileStorage: FileStorage{
			Type: "fs",
			Fs: struct {
				Path string
			}{Path: formModel.FileStoragePath},
		},
	}

	err = os.MkdirAll("/var/frogboard", 0744)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error creating config dir: %s\n", err.Error())))
		return
	}

	f, err := os.Create("/var/frogboard/config.toml")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error creatign config file: %s\n", err.Error())))
		return
	}
	toml.NewEncoder(f).Encode(&config)
	f.Close()

	w.Write([]byte(fmt.Sprintf("Setup complete. Restart application to start using FrogBoard.\nLogin:admin Password:%s\n", password)))
}
