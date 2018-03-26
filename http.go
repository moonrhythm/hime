package hime

import (
	"context"
	"net/http"
)

func (app *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, ctxKeyApp, app)
	r = r.WithContext(ctx)
	app.handler.ServeHTTP(w, r)
}

// ListenAndServe is the shotcut for http.ListenAndServe
func (app *app) ListenAndServe(addr string) (err error) {
	srv := http.Server{
		Addr:    addr,
		Handler: app,
	}

	return srv.ListenAndServe()
}
