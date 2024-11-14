package main

import (
	"net/http"

	"github.com/paluras/product-recall-system/internal/scraper"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	url := "https://www.ansvsa.ro/informatii-pentru-public/produse-rechemateretrase/"

	data, err := scraper.Scrape(url)
	if err != nil {
		app.errorLog.Printf("Scraping error: %v", err)
		http.Error(w, "Failed to scrape data", http.StatusInternalServerError)
		return
	}

	templateData := struct {
		Data []scraper.ScrapedData
	}{
		Data: data,
	}

	err = app.templates.ExecuteTemplate(w, "home.html", templateData)
	if err != nil {
		app.errorLog.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
