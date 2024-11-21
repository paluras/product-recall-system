package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/paluras/product-recall-system/internal/models"
	"github.com/paluras/product-recall-system/internal/scraper"
)

type Service struct {
	db       *models.DB
	logger   *slog.Logger
	timeout  time.Duration
	interval time.Duration
}

type ServiceConfig struct {
	DB       *models.DB
	Logger   *slog.Logger
	Interval time.Duration
	Timeout  time.Duration
}

func NewService(config ServiceConfig) *Service {
	return &Service{
		db:       config.DB,
		logger:   config.Logger,
		timeout:  config.Timeout,
		interval: config.Interval,
	}
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	if err := s.runScraping(ctx); err != nil {
		s.logger.Error("initial scraping failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.runScraping(ctx); err != nil {
				s.logger.Error("scraping failed", "error", err)
			}
		}
	}
}

func (s *Service) runScraping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	s.logger.Info("starting scrape")

	items, err := scraper.Scrape(ctx)
	if err != nil {
		return fmt.Errorf("scraping failed: %w", err)
	}

	s.logger.Info("scrape completed", "items_found", len(items))

	for _, data := range items {
		exists, err := s.db.ItemExists(data.Link)
		if err != nil {
			s.logger.Error("error checking item existence",
				"error", err,
				"title", data.Title)
			continue
		}

		if !exists {
			item := models.FromScraperData(data)
			if err := s.db.InsertItem(item); err != nil {
				s.logger.Error("error inserting item",
					"error", err,
					"title", data.Title)
				continue
			}
			s.logger.Info("stored new item", "title", data.Title)
		} else {
			s.logger.Debug("item already exists", "title", data.Title)
		}
	}

	return nil
}
