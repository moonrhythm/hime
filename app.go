package hime

import (
	"context"
	"crypto/tls"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App is the hime app
type App struct {
	srv     http.Server
	handler http.Handler
	routes  Routes
	globals Globals

	template      map[string]*tmpl
	templateFuncs []template.FuncMap

	gs *GracefulShutdown
}

var (
	ctxKeyApp = struct{}{}
)

// New creates new app
func New() *App {
	app := &App{}
	app.srv.Handler = app
	return app
}

// Clone clones app
func (app *App) Clone() *App {
	x := &App{
		srv: http.Server{
			Addr:              app.srv.Addr,
			ReadTimeout:       app.srv.ReadTimeout,
			ReadHeaderTimeout: app.srv.ReadHeaderTimeout,
			WriteTimeout:      app.srv.WriteTimeout,
			IdleTimeout:       app.srv.IdleTimeout,
			MaxHeaderBytes:    app.srv.MaxHeaderBytes,
			TLSNextProto:      cloneTLSNextProto(app.srv.TLSNextProto),
			ConnState:         app.srv.ConnState,
			ErrorLog:          app.srv.ErrorLog,
		},
		handler:       app.handler,
		routes:        cloneRoutes(app.routes),
		globals:       cloneGlobals(app.globals),
		template:      cloneTmpl(app.template),
		templateFuncs: cloneFuncMaps(app.templateFuncs),
	}
	x.srv.Handler = x

	if app.srv.TLSConfig != nil {
		x.srv.TLSConfig = app.srv.TLSConfig.Clone()
	}

	if app.gs != nil {
		x.gs = &GracefulShutdown{
			timeout: app.gs.timeout,
			wait:    app.gs.wait,
			notiFns: app.gs.notiFns,
		}
	}
	return x
}

// Address sets server address
func (app *App) Address(addr string) *App {
	app.srv.Addr = addr
	return app
}

// Handler sets the handler
func (app *App) Handler(h http.Handler) *App {
	app.handler = h
	return app
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, ctxKeyApp, app)
	r = r.WithContext(ctx)

	if app.handler == nil {
		http.DefaultServeMux.ServeHTTP(w, r)
		return
	}
	app.handler.ServeHTTP(w, r)
}

// Server returns server inside app
func (app *App) Server() *http.Server {
	return &app.srv
}

// Shutdown shutdowns server
func (app *App) Shutdown(ctx context.Context) error {
	if app.gs != nil {
		for _, fn := range app.gs.notiFns {
			go fn()
		}

		if app.gs.wait > 0 {
			time.Sleep(app.gs.wait)
		}

		if app.gs.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, app.gs.timeout)
			defer cancel()
		}
	}

	return app.srv.Shutdown(ctx)
}

func (app *App) listenAndServe() error {
	if app.srv.TLSConfig != nil {
		return app.srv.ListenAndServeTLS("", "")
	}

	return app.srv.ListenAndServe()
}

// ListenAndServe starts web server
func (app *App) ListenAndServe() error {
	if app.gs != nil {
		return app.startGracefulShutdown()
	}

	return app.listenAndServe()
}

func (app *App) ensureTLSConfig() {
	if app.srv.TLSConfig == nil {
		app.srv.TLSConfig = &tls.Config{}
	}
}

// TLS sets cert and key file
func (app *App) TLS(certFile, keyFile string) *App {
	app.ensureTLSConfig()

	err := loadTLSCertKey(app.srv.TLSConfig, certFile, keyFile)
	if err != nil {
		panicf("load key pair; %v", err)
	}

	return app
}

// SelfSign generates self sign cert
func (app *App) SelfSign(s SelfSign) *App {
	app.ensureTLSConfig()

	err := s.config(app.srv.TLSConfig)
	if err != nil {
		panicf("generate self sign; %v", err)
	}

	return app
}

// GracefulShutdown changes server to graceful shutdown mode
func (app *App) GracefulShutdown() *GracefulShutdown {
	if app.gs == nil {
		app.gs = &GracefulShutdown{}
	}

	return app.gs
}

func (app *App) startGracefulShutdown() error {
	errChan := make(chan error)

	go func() {
		if err := app.listenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-stop:
		return app.Shutdown(context.Background())
	}
}
