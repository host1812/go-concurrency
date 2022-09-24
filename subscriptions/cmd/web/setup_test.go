package main

import (
	"context"
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/host1812/go-concurrency/subscriptions/data"
)

var testApp Config

func TestMain(m *testing.M) {
	gob.Register(data.User{})
	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	testApp = Config{
		Session:       session,
		DB:            nil,
		InfoLog:       log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		ErrorLog:      log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		Wait:          &sync.WaitGroup{},
		ErrorChan:     make(chan error),
		ErrorChanDone: make(chan bool),
	}

	// dummy mailer
	errorChan := make(chan error)
	mailerChan := make(chan Message, 100)
	mailerDone := make(chan bool)

	testApp.Mailer = Mail{
		Wait:       testApp.Wait,
		ErrorChan:  errorChan,
		MailerChan: mailerChan,
		DoneChan:   mailerDone,
	}

	go func() {
		select {
		case <-testApp.Mailer.MailerChan:
		case <-testApp.Mailer.ErrorChan:
		case <-testApp.Mailer.DoneChan:
			return
		}
	}()

	go func() {
		for {
			select {
			case err := <-testApp.ErrorChan:
				testApp.ErrorLog.Println(err)
			case <-testApp.ErrorChanDone:
				return
			}
		}
	}()

	os.Exit(m.Run())
}

func getCtx(req *http.Request) context.Context {
	ctx, err := testApp.Session.Load(
		req.Context(),
		req.Header.Get("X-Session"),
	)
	if err != nil {
		log.Println(err)
	}
	return ctx
}
