package web

import (
	"net/http"
	"regexp"

	"github.com/paluras/product-recall-system/internal/models"
)

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	recalls, err := s.db.GetLatest20Items()
	if err != nil {
		s.logger.Error("error fetching items", "error", err)
		s.serverError(w, r, err)
		return
	}

	data := struct {
		Recalls []models.ScrapedItem
		Error   string
		Success string
	}{
		Recalls: recalls,
		Error:   s.session.PopString(r.Context(), "error"),
		Success: s.session.PopString(r.Context(), "success"),
	}

	err = s.templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		s.logger.Error("template error", "error", err)
		s.serverError(w, r, err)
		return
	}
}

func (s *Server) unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		s.logger.Error("no token provided")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := s.db.UnsubscribeWithToken(token)
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	data := struct {
		Success string
	}{
		Success: "You have been successfully unsubscribed.",
	}

	err = s.templates.ExecuteTemplate(w, "unsubscribe.html", data)
	if err != nil {
		s.serverError(w, r, err)
	}
}

func (s *Server) postSubscriber(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.session.Put(r.Context(), "error", "Error parsing the form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	email := r.PostForm.Get("subscribe")

	if !match(email, EmailRegex) {
		s.session.Put(r.Context(), "error", "Invalid email format")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	exists, err := s.db.EmailExists(email)
	if err != nil {
		s.session.Put(r.Context(), "error", "Server error")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if exists {
		s.session.Put(r.Context(), "error", "This email is already subscribed")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = s.db.AddSubscriber(email)
	if err != nil {
		s.session.Put(r.Context(), "error", "Error adding subscriber")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s.session.Put(r.Context(), "success", "Successfully subscribed!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) sitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	http.ServeFile(w, r, "ui/static/sitemap.xml")
}

func (s *Server) robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	http.ServeFile(w, r, "ui/static/robots.txt")
}
