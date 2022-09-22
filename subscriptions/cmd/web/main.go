package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const port = 80
const maxDbTries = 10

func main() {
	db := initDB()
	db.Ping()

	session := initSession()
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
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

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
