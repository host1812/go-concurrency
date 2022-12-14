package main

import (
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/host1812/go-concurrency/subscriptions/data"
)

var path = "./cmd/web/templates"

type TemplateData struct {
	StringMap     map[string]string
	IntMap        map[string]int
	FloatMap      map[string]float64
	Data          map[string]any
	Flash         string
	Warning       string
	Error         string
	Authenticated bool
	Now           time.Time
	User          *data.User
}

func (app *Config) render(
	w http.ResponseWriter,
	r *http.Request,
	t string,
	td *TemplateData,
) {
	partials := []string{
		fmt.Sprintf("%s/base.layout.gohtml", path),
		fmt.Sprintf("%s/header.partial.gohtml", path),
		fmt.Sprintf("%s/footer.partial.gohtml", path),
		fmt.Sprintf("%s/navbar.partial.gohtml", path),
		fmt.Sprintf("%s/alerts.partial.gohtml", path),
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("%s/%s", path, t))

	for _, x := range partials {
		templateSlice = append(templateSlice, x)
	}

	if td == nil {
		app.InfoLog.Println("initialize template data")
		td = &TemplateData{}
	}

	tmpl, err := template.ParseFiles(templateSlice...)
	if err != nil {
		app.ErrorLog.Println("error rendering template, err:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, app.AddDefaultData(td, r)); err != nil {
		app.ErrorLog.Println("error rendering template, err:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Config) AddDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	if app.IsAuthenticated(r) {
		td.Authenticated = true
		u, ok := app.Session.Get(r.Context(), "user").(data.User)
		if !ok {
			app.ErrorLog.Println("err getting user from session")
		} else {
			td.User = &u
		}
	} else {
		td.Authenticated = false
	}
	td.Now = time.Now()

	return td
}

func (app *Config) IsAuthenticated(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "userId")
}
