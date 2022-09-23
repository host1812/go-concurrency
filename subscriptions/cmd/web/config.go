package main

import (
	"database/sql"
	"log"
	"sync"

	"github.com/alexedwards/scs/v2"
	"github.com/host1812/go-concurrency/subscriptions/data"
)

type Config struct {
	Session  *scs.SessionManager
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	Wait     *sync.WaitGroup
	Models   data.Models
}

func (app *Config) GetInfoLogger() *log.Logger {
	return app.InfoLog
}
