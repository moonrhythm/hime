package hime

import (
	"context"
	"crypto/tls"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App is the hime app
type App struct {
	// Addr is server address
	Addr string

	// TLSConfig overrides http.Server TLSConfig
	TLSConfig *tls.Config

	// ReadTimeout overrides http.Server ReadTimeout
	ReadTimeout time.Duration

	// ReadHeaderTimeout overrides http.Server ReadHeaderTimeout
	ReadHeaderTimeout time.Duration

	// WriteTimeout overrides http.Server WriteTimeout
	WriteTimeout time.Duration

	// IdleTimeout overrides http.Server IdleTimeout
	IdleTimeout time.Duration

	// MaxHeaderBytes overrides http.Server MaxHeaderBytes
	MaxHeaderBytes int

	// TLSNextProto overrides http.Server TLSNextProto
	TLSNextProto map[string]func(*http.Server, *tls.Conn, http.Handler)

	// ConnState overrides http.Server ConnState
	ConnState func(net.Conn, http.ConnState)

	// ErrorLog overrides http.Server ErrorLog
	ErrorLog *log.Logger

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
	return &App{}
}

// Clone clones app
func (app *App) Clone() *App {
	x := &App{
		Addr:              app.Addr,
		ReadTimeout:       app.ReadTimeout,
		ReadHeaderTimeout: app.ReadHeaderTimeout,
		WriteTimeout:      app.WriteTimeout,
		IdleTimeout:       app.IdleTimeout,
		MaxHeaderBytes:    app.MaxHeaderBytes,
		TLSNextProto:      cloneTLSNextProto(app.TLSNextProto),
		ConnState:         app.ConnState,
		ErrorLog:          app.ErrorLog,
		handler:           app.handler,
		routes:            cloneRoutes(app.routes),
		globals:           cloneGlobals(app.globals),
		template:          cloneTmpl(app.template),
		templateFuncs:     cloneFuncMaps(app.templateFuncs),
	}
	if app.TLSConfig != nil {
		x.TLSConfig = app.TLSConfig.Clone()
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

func cloneTLSNextProto(xs map[string]func(*http.Server, *tls.Conn, http.Handler)) map[string]func(*http.Server, *tls.Conn, http.Handler) {
	if xs == nil {
		return nil
	}

	rs := make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	for k, v := range xs {
		rs[k] = v
	}
	return rs
}

// Address sets server address
func (app *App) Address(addr string) *App {
	app.Addr = addr
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

func (app *App) configServer() {
	app.srv.Addr = app.Addr
	app.srv.TLSConfig = app.TLSConfig
	app.srv.ReadTimeout = app.ReadTimeout
	app.srv.ReadHeaderTimeout = app.ReadHeaderTimeout
	app.srv.WriteTimeout = app.WriteTimeout
	app.srv.IdleTimeout = app.IdleTimeout
	app.srv.MaxHeaderBytes = app.MaxHeaderBytes
	app.srv.TLSNextProto = app.TLSNextProto
	app.srv.ConnState = app.ConnState
	app.srv.ErrorLog = app.ErrorLog
	app.srv.Handler = app
}

func (app *App) listenAndServe() error {
	app.configServer()

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
	if app.TLSConfig == nil {
		app.TLSConfig = &tls.Config{}
	}
}

// TLS sets cert and key file
func (app *App) TLS(certFile, keyFile string) *App {
	app.ensureTLSConfig()

	err := loadTLSCertKey(app.TLSConfig, certFile, keyFile)
	if err != nil {
		panicf("load key pair; %v", err)
	}

	return app
}

// SelfSign generates self sign cert
func (app *App) SelfSign(s SelfSign) *App {
	app.ensureTLSConfig()

	err := s.config(app.TLSConfig)
	if err != nil {
		panicf("generate self sign; %v", err)
	}

	return app
}

// OnShutdown calls server.RegisterOnShutdown(fn)
func (app *App) OnShutdown(fn func()) *App {
	app.srv.RegisterOnShutdown(fn)
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
		for _, fn := range app.gs.notiFns {
			go fn()
		}

		if app.gs.wait > 0 {
			time.Sleep(app.gs.wait)
		}

		ctx := context.Background()
		if app.gs.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, app.gs.timeout)
			defer cancel()
		}

		return app.srv.Shutdown(ctx)
	}
}
