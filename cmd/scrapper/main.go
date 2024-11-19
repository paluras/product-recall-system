package main

import (
	"log"

	"github.com/paluras/product-recall-system/configs"
	"github.com/paluras/product-recall-system/internal/models"
	"github.com/paluras/product-recall-system/internal/scraper"
)

func main() {
	conf := configs.ParseFlags()

	dsn := conf.DSN()
	db, err := models.NewDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Starting scrape...")

	items, err := scraper.Scrape()
	if err != nil {
		log.Fatal("Scraping failed:", err)
	}

	log.Printf("Found %d items", len(items))

	for _, data := range items {
		exists, err := db.ItemExists(data.Link)
		if err != nil {
			log.Printf("Error checking item existence: %v", err)
			continue
		}

		if !exists {
			item := models.FromScraperData(data)
			if err := db.InsertItem(item); err != nil {
				log.Printf("Error inserting item %s: %v", data.Title, err)
				continue
			}
			log.Printf("Stored new item: %s", data.Title)
		} else {
			log.Printf("Item already exists: %s", data.Title)
		}
	}

	log.Println("Scrape completed")
}
