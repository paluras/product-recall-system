package main

import (
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/paluras/product-recall-system/configs"
	"github.com/paluras/product-recall-system/internal/models"
)

type application struct {
	errorLog  *log.Logger
	infoLog   *log.Logger
	templates *template.Template
	db        *models.DB
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	conf := configs.ParseFlags()

	templates, err := template.ParseFiles("./ui/html/pages/home.html")
	if err != nil {
		errorLog.Fatal(err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		conf.DBUser, conf.DBPassword, conf.DBHost, conf.DBName)

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
	}

	err = app.serve()
	if err != nil {
		errorLog.Fatal(err)
	}
}
