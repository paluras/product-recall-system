package web

import (
	"net/http"
	"regexp"
)

func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.Error(err.Error(),
		"method", r.Method,
		"uri", r.URL.RequestURI(),
	)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func match(value string, rx *regexp.Regexp) bool {
	if value == "" {
		return false
	}
	return rx.MatchString(value)
}
