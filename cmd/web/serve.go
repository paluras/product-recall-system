package main

import (
	"net/http"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:    ":54321",
		Handler: app.routes(),
	}

	app.infoLog.Printf("Starting server on %s", srv.Addr)
	return srv.ListenAndServe()
}
