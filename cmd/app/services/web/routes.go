package web

import (
	"net/http"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("ui/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("GET /sitemap.xml", s.sitemap)
	mux.HandleFunc("GET /robots.txt", s.robots)

	mux.HandleFunc("GET /", s.home)
	mux.HandleFunc("POST /subscribe", s.postSubscriber)
	mux.HandleFunc("GET /unsubscribe", s.unsubscribe)

	return secureHeaders(mux)
}
