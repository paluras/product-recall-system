package main

import (
	"html/template"
	"log"
	"os"
)

type application struct {
	errorLog  *log.Logger
	infoLog   *log.Logger
	templates *template.Template
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	templates, err := template.ParseFiles("./ui/html/pages/home.html")
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:  errorLog,
		infoLog:   infoLog,
		templates: templates,
	}

	err = app.serve()
	if err != nil {
		errorLog.Fatal(err)
	}
}
