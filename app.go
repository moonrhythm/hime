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

	srv          http.Server
	handler      http.Handler
	routes       Routes
	globals      Globals
	beforeRender middleware.Middleware

	template      map[string]*tmpl
	templateFuncs []template.FuncMap

	gracefulShutdown *gracefulShutdown

	certFile, keyFile string
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
		beforeRender:      app.beforeRender,
		template:          cloneTmpl(app.template),
		templateFuncs:     cloneFuncMaps(app.templateFuncs),
		gracefulShutdown:  &*app.gracefulShutdown,
		certFile:          app.certFile,
		keyFile:           app.keyFile,
	}
	if app.TLSConfig != nil {
		x.TLSConfig = app.TLSConfig.Clone()
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

	return app.srv.ListenAndServe()
}

func (app *App) listenAndServeTLS(certFile, keyFile string) error {
	app.configServer()

	return app.srv.ListenAndServeTLS(certFile, keyFile)
}

// ListenAndServe starts web server
func (app *App) ListenAndServe() error {
	if app.certFile != "" && app.keyFile != "" {
		return app.ListenAndServeTLS(app.certFile, app.keyFile)
	}

	if app.gracefulShutdown != nil {
		return app.GracefulShutdown().ListenAndServe()
	}

	return app.listenAndServe()
}

// TLS sets cert and key file
func (app *App) TLS(certFile, keyFile string) *App {
	app.certFile, app.keyFile = certFile, keyFile
	return app
}

// ListenAndServeTLS starts web server in tls mode
func (app *App) ListenAndServeTLS(certFile, keyFile string) error {
	if app.gracefulShutdown != nil {
		return app.GracefulShutdown().ListenAndServeTLS(certFile, keyFile)
	}

	return app.listenAndServeTLS(certFile, keyFile)
}

// GracefulShutdown returns graceful shutdown server
func (app *App) GracefulShutdown() *GracefulShutdownApp {
	if app.gracefulShutdown == nil {
		app.gracefulShutdown = &gracefulShutdown{}
	}

	return &GracefulShutdownApp{
		App:              app,
		gracefulShutdown: app.gracefulShutdown,
	}
}
