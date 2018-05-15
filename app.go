package hime

import (
	"context"
	"crypto/tls"
	"html/template"
	"log"
	"mime"
	"net"
	"net/http"
	"time"

	"github.com/acoshift/middleware"
)

// App is the hime app
type App struct {
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

	srv          http.Server
	handler      http.Handler
	routes       Routes
	globals      Globals
	beforeRender middleware.Middleware

	template     map[string]*tmpl
	templateFunc []template.FuncMap

	graceful struct {
		timeout time.Duration
		wait    time.Duration
	}
}

var (
	ctxKeyApp = struct{}{}
)

func init() {
	mime.AddExtensionType(".js", "text/javascript")
}

// New creates new app
func New() *App {
	return &App{}
}

// Handler sets the handler
func (app *App) Handler(h http.Handler) *App {
	app.handler = h
	return app
}

// BeforeRender runs given middleware for before render,
// ex. View, JSON, String, Bytes, CopyForm, etc
func (app *App) BeforeRender(m middleware.Middleware) *App {
	app.beforeRender = m
	return app
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, ctxKeyApp, app)
	r = r.WithContext(ctx)
	app.handler.ServeHTTP(w, r)
}

func (app *App) configServer(addr string) {
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
	app.srv.Addr = addr
}

// ListenAndServe starts web server
func (app *App) ListenAndServe(addr string) error {
	app.configServer(addr)

	return app.srv.ListenAndServe()
}

// ListenAndServeTLS starts web server in tls mode
func (app *App) ListenAndServeTLS(addr, certFile, keyFile string) error {
	app.configServer(addr)

	return app.srv.ListenAndServeTLS(certFile, keyFile)
}

// GracefulShutdown returns graceful shutdown server
func (app *App) GracefulShutdown() *GracefulShutdown {
	return &GracefulShutdown{
		App:     app,
		timeout: app.graceful.timeout,
		wait:    app.graceful.wait,
	}
}
