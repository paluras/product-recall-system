package main

import (
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/paluras/product-recall-system/configs"
	"github.com/paluras/product-recall-system/internal/models"
	"github.com/paluras/product-recall-system/internal/notify"
)

type application struct {
	errorLog     *log.Logger
	infoLog      *log.Logger
	templates    *template.Template
	db           *models.DB
	session      *scs.SessionManager
	logger       *slog.Logger
	emailService *notify.EmailService
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	conf := configs.ParseFlags()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	templates, err := template.ParseFiles(
		"./ui/html/pages/home.html",
		"./ui/html/pages/unsubscribe.html")
	if err != nil {
		logger.Error("Template error")
	}

	dsn := conf.DSN()

	db, err := models.NewDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	resendKey := os.Getenv("RESEND_API_KEY")
	var emailService *notify.EmailService
	if resendKey != "" {
		emailService, err = notify.NewEmailService(notify.EmailConfig{
			APIKey:    resendKey,
			FromEmail: "Latest Alert <alert@latest.produseretrase.eu>",
		}, db)
		if err != nil {
			errorLog.Printf("Failed to initialize email service: %v", err)
		}
	}

	app := &application{
		errorLog:     errorLog,
		infoLog:      infoLog,
		templates:    templates,
		db:           db,
		session:      session,
		logger:       logger,
		emailService: emailService,
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
