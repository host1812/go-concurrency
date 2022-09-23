package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/host1812/go-concurrency/subscriptions/data"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const port = 80
const maxDbTries = 10

var app *Config

func main() {
	db := initDB()
	db.Ping()

	session := initSession()

	//loggers
	infoLog := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	wg := sync.WaitGroup{}

	app = &Config{
		Session:  session,
		DB:       db,
		Wait:     &wg,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
		Models:   data.New(db),
	}
	go app.listenForShutdown()

	app.Mailer = app.createMail()

	go app.listenForMail()

	app.serve()
}

// start http server
func (app *Config) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: app.routes(),
	}
	app.InfoLog.Printf("starting web server on port %d\n", port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Panicf("failed to start server on port %d, err: %s\n", port, err)
	}
}

func initDB() *sql.DB {
	conn := connectToDB()
	if conn == nil {
		log.Panic("all attempts to connect to the database failed!")
	}
	return conn
}

func connectToDB() *sql.DB {
	try := 0

	dsn := os.Getenv("DSN")

	for try < maxDbTries {
		conn, err := openDB(dsn)
		if err != nil {
			log.Println("failed to connect to the database, retrying...")
			try++
		} else {
			return conn
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func initSession() *scs.SessionManager {
	gob.Register(data.User{})
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	return session
}

func initRedis() *redis.Pool {
	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}
	return redisPool
}

func (app *Config) listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.shutdown()
	os.Exit(0)
}

func (app *Config) shutdown() {
	// perform cleanup
	app.InfoLog.Println("doing cleanup")

	// block until waitgroup is empty
	app.Wait.Wait()

	// stopping mailer
	app.Mailer.DoneChan <- true

	app.InfoLog.Println("closing channels")
	close(app.Mailer.MailerChan)
	close(app.Mailer.ErrorChan)
	close(app.Mailer.DoneChan)

}

func (app *Config) createMail() Mail {
	errChan := make(chan error)
	mailerChan := make(chan Message, 100)
	doneChan := make(chan bool)

	m := Mail{
		Domain:      "localhost",
		Host:        "localhost",
		Port:        1025,
		Encryption:  "none",
		FromAddress: "info@info.com",
		FromName:    "info@info.com",
		Wait:        app.Wait,
		ErrorChan:   errChan,
		MailerChan:  mailerChan,
		DoneChan:    doneChan,
	}
	return m
}
