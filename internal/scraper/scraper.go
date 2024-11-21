package scraper

import (
	"context"
	"fmt"
	"net/http"
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

func Scrape(ctx context.Context) ([]ScrapedData, error) {
	client := utils.CreateHTTPClient()

	req, err := http.NewRequestWithContext(ctx, "GET", ScraperURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var results []ScrapedData

	doc.Find(".pt-cv-ifield").Each(func(i int, s *goquery.Selection) {

		select {
		case <-ctx.Done():
			return
		default:
		}

		title := s.Find(".pt-cv-title a").Text()
		link, exists := s.Find(".pt-cv-title a").Attr("href")
		if !exists {
			return
		}

		date := s.Find(".entry-date time").Text()
		parsedDate, err := time.Parse("02/01/2006", date)
		if err != nil {
			return
		}

		results = append(results, ScrapedData{
			Title: title,
			Link:  link,
			Date:  parsedDate,
		})
	})

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Date.After(results[j].Date)
	})

	return results, nil
}
