package web

import (
	"net/http"
	"time"

	"github.com/paluras/product-recall-system/internal/middleware"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("ui/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("GET /sitemap.xml", s.sitemap)
	mux.HandleFunc("GET /robots.txt", s.robots)

	mux.HandleFunc("GET /", s.home)

	limiter := middleware.NewRateLimiter(2, time.Minute, s.logger)

	mux.Handle("POST /subscribe", limiter.Limit(http.HandlerFunc(s.postSubscriber)))
	mux.HandleFunc("GET /confirm", s.confirmSubscription)

	mux.HandleFunc("GET /unsubscribe", s.unsubscribe)

	return secureHeaders(mux)
}
