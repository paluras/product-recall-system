package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/paluras/product-recall-system/cmd/app/services/notifier"
	"github.com/paluras/product-recall-system/cmd/app/services/scraper"
	"github.com/paluras/product-recall-system/cmd/app/services/web"
	"github.com/paluras/product-recall-system/configs"
	"github.com/paluras/product-recall-system/internal/models"
	"github.com/paluras/product-recall-system/internal/notify"
)

type App struct {
	webServer *web.Server
	scraper   *scraper.Service
	notifier  *notifier.Service
	logger    *slog.Logger
	config    *configs.Config
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := configs.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	db, err := models.NewDB(cfg.Database)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	emailSvc, err := notify.NewEmailService(notify.AWSConfig{
		Region:    cfg.AWS.Region,
		FromEmail: cfg.Email.FromEmail,
	}, db)
	if err != nil {
		logger.Error("failed to initialize email service", "error", err)
		os.Exit(1)
	}

	templates, err := template.ParseGlob(cfg.Server.TemplatesDir + "/**/*.html")
	if err != nil {
		logger.Error("failed to parse templates", "error", err)
		os.Exit(1)
	}

	app := &App{
		webServer: web.NewServer(web.ServerConfig{
			DB:        db,
			Logger:    logger,
			Addr:      cfg.Server.Host + ":" + cfg.Server.Port,
			Templates: templates,
			EmailSvc:  emailSvc,
		}),
		scraper: scraper.NewService(scraper.ServiceConfig{
			DB:       db,
			Logger:   logger,
			Interval: cfg.Scraper.Interval,
			Timeout:  cfg.Scraper.Timeout,
		}),
		notifier: notifier.NewService(notifier.ServiceConfig{
			DB:        db,
			EmailSvc:  emailSvc,
			Logger:    logger,
			Interval:  cfg.Notifier.Interval,
			BatchSize: cfg.Notifier.BatchSize,
		}),
		logger: logger,
		config: cfg,
	}

	if err := app.Run(context.Background()); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.webServer.Run(ctx); err != nil {
			a.logger.Error("web server error", "error", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.scraper.Run(ctx); err != nil {
			a.logger.Error("scraper error", "error", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.notifier.Run(ctx); err != nil {
			a.logger.Error("notifier error", "error", err)
			cancel()
		}
	}()

	return a.handleShutdown(ctx, cancel, &wg)
}

func (a *App) handleShutdown(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		a.logger.Info("context cancelled, shutting down")
	case sig := <-sigChan:
		a.logger.Info("shutdown signal received", "signal", sig)
		cancel()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("graceful shutdown completed")
		return nil
	case <-time.After(a.config.Server.ShutdownTimeout):
		return fmt.Errorf("shutdown timeout after %v", a.config.Server.ShutdownTimeout)
	}
}
