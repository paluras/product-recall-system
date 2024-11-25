package web

import (
	"database/sql"
	"net/http"
	"regexp"
	"strings"

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

	// Check both subscribers and pending tables
	subscriberExists, err := s.db.EmailExists(email)
	if err != nil {
		s.session.Put(r.Context(), "error", "Server error")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pendingExists, err := s.db.PendingEmailExists(email)
	if err != nil {
		s.session.Put(r.Context(), "error", "Server error")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if subscriberExists {
		s.session.Put(r.Context(), "error", "This email is already subscribed")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if pendingExists {
		s.session.Put(r.Context(), "error", "Verification email already sent. Please check your inbox")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Create pending subscription
	token, err := s.db.CreatePendingSubscriber(email)
	if err != nil {
		s.session.Put(r.Context(), "error", "Failed to create subscription")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Send verification email
	if err := s.emailSvc.SendVerificationEmail(email, token); err != nil {
		// Clean up pending subscription if email fails
		if cleanErr := s.db.DeletePendingSubscriber(email); cleanErr != nil {
			s.logger.Error("failed to clean up pending subscriber", "error", cleanErr)
		}
		s.logger.Error("failed to send confirmation email", "error", err)
		s.session.Put(r.Context(), "error", "Failed to send confirmation email")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s.session.Put(r.Context(), "success", "Please check your email to confirm your subscription")
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

func (s *Server) confirmSubscription(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	if err := s.db.ConfirmSubscriber(token); err != nil {
		s.logger.Error("failed to confirm subscription",
			"error", err,
			"token", token)

		if err == sql.ErrNoRows {
			http.Error(w, "Token not found or expired", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "already subscribed") {
			http.Error(w, "Email already confirmed", http.StatusConflict)
			return
		}

		http.Error(w, "Failed to confirm subscription", http.StatusInternalServerError)
		return
	}

	s.session.Put(r.Context(), "success", "Your subscription has been confirmed successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
