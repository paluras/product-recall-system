package web

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/paluras/product-recall-system/internal/models"
)

type Server struct {
	db        *models.DB
	logger    *slog.Logger
	session   *scs.SessionManager
	templates *template.Template
	addr      string
}

type ServerConfig struct {
	DB        *models.DB
	Logger    *slog.Logger
	Addr      string
	Templates *template.Template
}

func NewServer(config ServerConfig) *Server {
	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return &Server{
		db:        config.DB,
		logger:    config.Logger,
		session:   session,
		templates: config.Templates,
		addr:      config.Addr,
	}
}

func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.session.LoadAndSave(s.routes()),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	serverError := make(chan error, 1)
	go func() {
		s.logger.Info("starting server", "addr", srv.Addr)
		serverError <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		s.logger.Info("shutting down server")
		return srv.Shutdown(shutdownCtx)
	}
}
