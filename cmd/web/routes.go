package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.home)
	mux.HandleFunc("POST /subscribe", app.PostSubscriber)
	mux.HandleFunc("GET /unsubscribe", app.unsubscribe)

	return mux
}
