package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const port = 80
const maxDbTries = 10

func main() {
	db := initDB()
	db.Ping()
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
