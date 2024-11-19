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
)

type application struct {
	errorLog  *log.Logger
	infoLog   *log.Logger
	templates *template.Template
	db        *models.DB
	session   *scs.SessionManager
	logger    *slog.Logger
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
	session.Cookie.Secure = false // set to true in production

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

	app := &application{
		errorLog:  errorLog,
		infoLog:   infoLog,
		templates: templates,
		db:        db,
		session:   session,
		logger:    logger,
	}

	subscribers, err := app.db.GetSubscribersMail()
	if err != nil {
		logger.Info("No subscribers")
	}

	for _, subscriber := range subscribers {
		logger.Info("Subscribers=", "subscribers", subscriber)
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
