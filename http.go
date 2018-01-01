package hime

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	if !app.gracefulShutdown {
		return srv.ListenAndServe()
	}

	serverCtx, cancelServer := context.WithCancel(context.Background())
	defer cancelServer()
	go func() {
		if err = srv.ListenAndServe(); err != http.ErrServerClosed {
			cancelServer()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	select {
	case <-serverCtx.Done():
		return
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), app.shutdownTimeout)
		defer cancel()
		err = srv.Shutdown(ctx)
	}
	return
}
