package hime

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdownApp

type gracefulShutdownApp struct {
	*app
	timeout   time.Duration
	wait      time.Duration
	notiFns   []func()
	beforeFns []func()
}

// ShutdownTimeout sets shutdown timeout for graceful shutdown
func (app *gracefulShutdownApp) Timeout(d time.Duration) GracefulShutdownApp {
	app.timeout = d
	return app
}

func (app *gracefulShutdownApp) Wait(d time.Duration) GracefulShutdownApp {
	app.wait = d
	return app
}

func (app *gracefulShutdownApp) Notify(fn func()) GracefulShutdownApp {
	if fn != nil {
		app.notiFns = append(app.notiFns, fn)
	}
	return app
}

func (app *gracefulShutdownApp) Before(fn func()) GracefulShutdownApp {
	if fn != nil {
		app.beforeFns = append(app.beforeFns, fn)
	}
	return app
}

// ListenAndServe is the shotcut for http.ListenAndServe
func (app *gracefulShutdownApp) ListenAndServe(addr string) (err error) {
	if app.srv == nil {
		app.srv = &http.Server{
			Addr:    addr,
			Handler: app,
		}
	}

	serverCtx, cancelServer := context.WithCancel(context.Background())
	defer cancelServer()
	go func() {
		if err = app.srv.ListenAndServe(); err != http.ErrServerClosed {
			cancelServer()
		}
	}()

	for _, fn := range app.beforeFns {
		fn()
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	select {
	case <-serverCtx.Done():
		return
	case <-stop:
		for _, fn := range app.notiFns {
			fn()
		}
		if app.wait > 0 {
			time.Sleep(app.wait)
		}
		ctx, cancel := context.WithTimeout(context.Background(), app.timeout)
		defer cancel()
		err = app.srv.Shutdown(ctx)
	}
	return
}
