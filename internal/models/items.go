package models

import (
	"time"

	"github.com/paluras/product-recall-system/internal/scraper"
)

type ScrapedItem struct {
	ID        int
	Title     string
	Link      string
	Date      time.Time
	CreatedAt time.Time
}

func (db *DB) GetLatest20Items() ([]ScrapedItem, error) {
	query := `
        SELECT id, title, link, date, created_at
        FROM scraped_items
        ORDER BY date DESC
        LIMIT 20
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ScrapedItem
	for rows.Next() {
		var item ScrapedItem
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Link,
			&item.Date,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func FromScraperData(data scraper.ScrapedData) ScrapedItem {
	return ScrapedItem{
		Title: data.Title,
		Link:  data.Link,
		Date:  data.Date,
	}
}

func (db *DB) InsertItem(item ScrapedItem) error {
	query := `
        INSERT INTO scraped_items (title, link, date)
        VALUES (?, ?, ?)
    `
	_, err := db.Exec(query, item.Title, item.Link, item.Date)
	return err
}

func (db *DB) ItemExists(link string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM scraped_items WHERE link = ?)"
	err := db.QueryRow(query, link).Scan(&exists)
	return exists, err
}

func (db *DB) GetLatestItems(limit int) ([]ScrapedItem, error) {
	query := `
        SELECT id, title, link, date, created_at
        FROM scraped_items
        ORDER BY date DESC
        LIMIT ?
    `

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ScrapedItem
	for rows.Next() {
		var item ScrapedItem
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Link,
			&item.Date,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
