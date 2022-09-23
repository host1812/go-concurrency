package main

import (
	"fmt"
	"net/http"
)

func (app *Config) Home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.gohtml", nil)
}

func (app *Config) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.RenewToken(r.Context())
	// parse form
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
	}
	// get email, password
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.InfoLog.Printf("%s - user not found\n", email)
		app.Session.Put(r.Context(), "error", "invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// check password
	valid, err := user.PasswordMatches(password)
	if err != nil {
		app.InfoLog.Printf("%s - not able to match password\n", email)
		app.Session.Put(r.Context(), "error", "invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !valid {
		msg := Message{
			To:      email,
			Subject: "Failed login attempt",
			Data:    fmt.Sprintf("Failed login attempt detected for %s", user.Email),
		}
		app.sendEmail(msg)
		app.InfoLog.Printf("%s - tried to authenticate with invalid password\n", email)
		app.Session.Put(r.Context(), "error", "invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// login successful
	app.InfoLog.Printf("%s - successfully authenticated\n", email)
	app.Session.Put(r.Context(), "userId", user.ID)
	app.Session.Put(r.Context(), "user", user)
	app.Session.Put(r.Context(), "flash", "successful login")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {
	// cleanup session
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) RegisterPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "register.page.gohtml", nil)
}

func (app *Config) PostRegisterPage(w http.ResponseWriter, r *http.Request) {
}

func (app *Config) ActivateAccount(w http.ResponseWriter, r *http.Request) {
}
