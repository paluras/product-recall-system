package main

import (
	"net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	data, err := app.db.GetLatest20Items()
	if err != nil {
		app.errorLog.Printf("Error fetching items: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = app.templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		app.errorLog.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
