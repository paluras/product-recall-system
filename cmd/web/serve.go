package main

import (
	"net/http"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:    ":54321",
		Handler: app.session.LoadAndSave(app.routes()),
	}

	app.logger.Info("Starting server ", "addr", srv.Addr)
	return srv.ListenAndServe()
}
