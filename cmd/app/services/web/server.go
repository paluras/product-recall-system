package web

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/paluras/product-recall-system/internal/models"
	"github.com/paluras/product-recall-system/internal/notify"
)

type Server struct {
	db        *models.DB
	logger    *slog.Logger
	session   *scs.SessionManager
	templates *template.Template
	addr      string
	emailSvc  *notify.EmailService
}

type ServerConfig struct {
	DB        *models.DB
	Logger    *slog.Logger
	Addr      string
	Templates *template.Template
	EmailSvc  *notify.EmailService
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
		emailSvc:  config.EmailSvc,
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

// func (s *Server) cleanupRoutine(ctx context.Context) {
// 	ticker := time.NewTicker(24 * time.Hour)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-ticker.C:
// 			if affected, err := s.db.DeleteExpiredPendingSubscribers(); err != nil {
// 				s.logger.Error("failed to cleanup expired verifications", "error", err)
// 			} else if affected > 0 {
// 				s.logger.Info("cleaned up expired verifications", "count", affected)
// 			}
// 		}
// 	}
// }
