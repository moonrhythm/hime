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
	notiChans []chan<- struct{}
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

func (app *gracefulShutdownApp) Notify(ch chan<- struct{}) GracefulShutdownApp {
	app.notiChans = append(app.notiChans, ch)
	return app
}

// ListenAndServe is the shotcut for http.ListenAndServe
func (app *gracefulShutdownApp) ListenAndServe(addr string) (err error) {
	srv := http.Server{
		Addr:    addr,
		Handler: app,
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
		for _, ch := range app.notiChans {
			ch <- struct{}{}
		}
		if app.wait > 0 {
			time.Sleep(app.wait)
		}
		ctx, cancel := context.WithTimeout(context.Background(), app.timeout)
		defer cancel()
		err = srv.Shutdown(ctx)
	}
	return
}
