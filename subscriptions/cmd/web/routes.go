package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	//middleware
	mux.Use(middleware.Recoverer)
	mux.Use(app.SessionLoad)

	//define
	mux.Get("/", app.Home)
	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLoginPage)
	mux.Get("/logout", app.Logout)
	mux.Get("/register", app.RegisterPage)
	mux.Post("/register", app.PostRegisterPage)
	mux.Get("/activate", app.ActivateAccount)

	mux.Get("/test-email", func(w http.ResponseWriter, r *http.Request) {
		m := Mail{
			Domain:      "localhost",
			Host:        "localhost",
			Port:        1025,
			Encryption:  "none",
			FromAddress: "info@company.com",
			FromName:    "info@company.com",
			ErrorChan:   make(chan error),
		}
		msg := Message{
			To:      "me@here.com",
			Subject: "Test email",
			Data:    "Hello",
		}
		app.InfoLog.Println("trying to send email")
		m.sendMail(msg, make(chan error))
		app.InfoLog.Println("email was sent successfully")
	})
	return mux
}
