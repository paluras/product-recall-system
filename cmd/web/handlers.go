package main

import (
	"net/http"
	"regexp"

	"github.com/paluras/product-recall-system/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	recalls, err := app.db.GetLatest20Items()
	if err != nil {
		app.errorLog.Printf("Error fetching items: %v", err)
		app.serverError(w, r, err)
		return
	}

	data := struct {
		Recalls []models.ScrapedItem
		Error   string
		Success string
	}{
		Recalls: recalls,
		Error:   app.session.PopString(r.Context(), "error"),
		Success: app.session.PopString(r.Context(), "success"),
	}

	err = app.templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		app.errorLog.Printf("Template error: %v", err)
		app.serverError(w, r, err)
		return
	}
}

func (app *application) unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		app.logger.Error("No token")
		return
	}

	err := app.db.UnsubscribeWithToken(token)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := struct {
		Success string
	}{
		Success: "You have been successfully unsubscribed.",
	}

	err = app.templates.ExecuteTemplate(w, "unsubscribe.html", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) PostSubscriber(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.session.Put(r.Context(), "error", "Error parsing the form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// honeypot: bots fill hidden fields, real users don't
	if r.PostForm.Get("website") != "" {
		app.session.Put(r.Context(), "success", "Verificați emailul pentru a confirma abonarea!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !app.limiter.allow(realIP(r)) {
		app.session.Put(r.Context(), "error", "Prea multe încercări. Vă rugăm așteptați înainte de a încerca din nou.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	email := r.PostForm.Get("subscribe")

	if !Match(email, EmailRegex) {
		app.session.Put(r.Context(), "error", "Invalid email format")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	exists, err := app.db.EmailExists(email)
	if err != nil {
		app.session.Put(r.Context(), "error", "Server error")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if exists {
		app.session.Put(r.Context(), "error", "This email is already subscribed")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	confirmToken, err := app.db.AddSubscriber(email)
	if err != nil {
		app.session.Put(r.Context(), "error", "Error adding subscriber")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if app.emailService != nil {
		go func() {
			if err := app.emailService.SendConfirmationEmail(email, confirmToken); err != nil {
				app.errorLog.Printf("Failed to send confirmation email to %s: %v", email, err)
			}
		}()
	}

	app.session.Put(r.Context(), "success", "Verificați emailul pentru a confirma abonarea!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) confirmSubscriber(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err := app.db.ConfirmSubscriber(token)
	if err != nil {
		app.logger.Error("confirmation failed", "err", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	data := struct {
		Success string
	}{
		Success: "Email confirmat cu succes! Vei primi notificări despre retragerile de produse.",
	}

	err = app.templates.ExecuteTemplate(w, "unsubscribe.html", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func Match(value string, rx *regexp.Regexp) bool {
	if value == "" {
		return false
	}
	return rx.MatchString(value)
}
