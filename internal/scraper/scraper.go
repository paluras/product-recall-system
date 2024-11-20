package scraper

import (
	"log"
	"sort"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/paluras/product-recall-system/internal/utils"
)

type ScrapedData struct {
	Title string
	Link  string
	Date  time.Time
}

const ScraperURL = "https://www.ansvsa.ro/informatii-pentru-public/produse-rechemateretrase/"

func Scrape() ([]ScrapedData, error) {

	client := utils.CreateHTTPClient()

	resp, err := client.Get(ScraperURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []ScrapedData

	doc.Find(".pt-cv-ifield").Each(func(i int, s *goquery.Selection) {

		title := s.Find(".pt-cv-title a").Text()
		link, _ := s.Find(".pt-cv-title a").Attr("href")
		date := s.Find(".entry-date time").Text()

		parsedDate, err := time.Parse("02/01/2006", date)
		if err != nil {
			log.Printf("Failed to parse date %q: %v", date, err)
			parsedDate = time.Time{}

		}

		results = append(results, ScrapedData{
			Title: title,
			Link:  link,
			Date:  parsedDate,
		})
	})

	sort.Slice(results, func(i, j int) bool {
		return results[i].Date.After(results[j].Date)
	})

	return results, nil
}
