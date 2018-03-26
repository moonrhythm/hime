package hime

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acoshift/middleware"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

// TemplateRoot sets template root to select when load
func (app *app) TemplateRoot(name string) App {
	app.templateRoot = name
	return app
}

// TemplateDir sets template dir
func (app *app) TemplateDir(path string) App {
	app.templateDir = path
	return app
}

// Handler sets app handler
func (app *app) Handler(factory HandlerFactory) App {
	app.handler = factory(app)
	return app
}

// Minify sets app minifier
func (app *app) Minify() App {
	app.minifier = minify.New()
	app.minifier.AddFunc("text/html", html.Minify)
	app.minifier.AddFunc("text/css", css.Minify)
	app.minifier.AddFunc("text/javascript", js.Minify)
	return app
}

func (app *app) BeforeRender(m middleware.Middleware) App {
	app.beforeRender = m
	return app
}

// GracefulShutdownApp

type gracefulShutdownApp struct {
	*app
	timeout time.Duration
	wait    time.Duration
}

// GracefulShutdown sets graceful shutdown to true
func (app *app) GracefulShutdown() GracefulShutdownApp {
	return &gracefulShutdownApp{
		app:     app,
		timeout: defShutdownTimeout,
	}
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
		if app.wait > 0 {
			time.Sleep(app.wait)
		}
		ctx, cancel := context.WithTimeout(context.Background(), app.timeout)
		defer cancel()
		err = srv.Shutdown(ctx)
	}
	return
}
