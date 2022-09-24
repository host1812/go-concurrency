package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/host1812/go-concurrency/subscriptions/data"
	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
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
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println("err parsing post form:", err)
	}

	// todo: data validation
	u := data.User{
		Email:     r.Form.Get("email"),
		FirstName: r.Form.Get("first-name"),
		LastName:  r.Form.Get("last-name"),
		Password:  r.Form.Get("password"),
		Active:    0,
		IsAdmin:   0,
	}

	_, err = u.Insert(u)
	if err != nil {
		app.ErrorLog.Println("err inserting user into db:", err)
		app.Session.Put(
			r.Context(),
			"error",
			fmt.Sprintln("err inserting user into db:", err),
		)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// send activation email
	url := fmt.Sprintf("http://localhost:80/activate?email=%s", u.Email)
	signedURL := GenerateTokenFromString(url)
	app.InfoLog.Println("signed url:", signedURL)
	msg := Message{
		To:       u.Email,
		Subject:  "Activate your account",
		Template: "confirmation-email",
		Data:     template.HTML(signedURL),
	}
	app.sendEmail(msg)
	app.Session.Put(r.Context(), "flash", "Registration successful. Check your email.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	// validate url
	url := r.RequestURI
	testURL := fmt.Sprintf("http://localhost:80%s", url)
	ok := VerifyToken(testURL)
	if !ok {
		app.Session.Put(r.Context(), "error", "Invalid URL.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// activate account
	u, err := app.Models.User.GetByEmail(r.URL.Query().Get("email"))
	if err != nil {
		app.Session.Put(r.Context(), "error", "No user found.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	u.Active = 1
	err = u.Update()
	if err != nil {
		app.Session.Put(r.Context(), "error", "Failed to update user.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "flash", "User activated.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)

}

func (app *Config) SubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	// get id of the plan
	id := r.URL.Query().Get("id")
	planID, _ := strconv.Atoi(id)
	plan, err := app.Models.Plan.GetOne(planID)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Failed to get plan from db.")
		http.Redirect(w, r, "/members/plans", http.StatusSeeOther)
		return
	}

	user, ok := app.Session.Get(r.Context(), "user").(data.User)
	if !ok {
		app.Session.Put(r.Context(), "error", "Login first!")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// generate invoice and email it
	app.Wait.Add(1)
	go func() {
		defer app.Wait.Done()

		invoice, err := app.getInvoice(user, plan)
		if err != nil {
			app.ErrorChan <- err
		}

		msg := Message{
			To:       user.Email,
			Subject:  "Your invoice",
			Data:     invoice,
			Template: "invoice",
		}

		app.sendEmail(msg)
	}()

	// generate manual
	app.Wait.Add(1)
	go func() {
		defer app.Wait.Done()

		pdf := app.generateManual(user, plan)
		err := pdf.OutputFileAndClose(fmt.Sprintf("./tmp/%d_manual.pdf", user.ID))
		if err != nil {
			app.ErrorChan <- err
			return
		}

		msg := Message{
			To:      user.Email,
			Subject: "Your manual",
			Data:    "Your user manual is attached",
			AttachmentMap: map[string]string{
				"manual.pdf": fmt.Sprintf("./tmp/%d_manual.pdf", user.ID),
			},
		}

		// send an email with manual
		app.sendEmail(msg)

		// test errors
		app.ErrorChan <- errors.New("some new error")
	}()

	// update db
	err = app.Models.Plan.SubscribeUserToPlan(user, *plan)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Error updating db")
		http.Redirect(w, r, "/members/plan", http.StatusSeeOther)
		return
	}

	u, err := app.Models.User.GetOne(user.ID)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Error updating db")
		http.Redirect(w, r, "/members/plan", http.StatusSeeOther)
		return
	}
	app.Session.Put(r.Context(), "user", u)

	// redirect
	app.Session.Put(r.Context(), "flash", "Subscribed")
	http.Redirect(w, r, "/members/plans", http.StatusSeeOther)
}

func (app *Config) getInvoice(u data.User, plan *data.Plan) (string, error) {
	app.InfoLog.Printf("plan amount formatted for plan id %d: %s", plan.ID, plan.PlanAmountFormatted)
	return plan.PlanAmountFormatted, nil
}

func (app *Config) generateManual(u data.User, p *data.Plan) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10)
	importer := gofpdi.NewImporter()

	// just do work
	time.Sleep(5 * time.Second)

	t := importer.ImportPage(pdf, "./pdf/manual.pdf", 1, "/MediaBox")
	pdf.AddPage()
	importer.UseImportedTemplate(pdf, t, 0, 0, 215.9, 0)

	pdf.SetX(75)
	pdf.SetY(150)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 4, fmt.Sprintf("%s %s", u.FirstName, u.LastName), "", "C", false)
	pdf.Ln(5)
	pdf.MultiCell(0, 4, fmt.Sprintf("%s User Guide", p.PlanName), "", "C", false)

	return pdf
}

func (app *Config) ChooseSubscription(w http.ResponseWriter, r *http.Request) {
	plans, err := app.Models.Plan.GetAll()
	if err != nil {
		app.ErrorLog.Println("err getting plans from db:", err)
		return
	}
	dataMap := make(map[string]any)
	dataMap["plans"] = plans
	app.render(w, r, "plans.page.gohtml", &TemplateData{
		Data: dataMap,
	})
}
