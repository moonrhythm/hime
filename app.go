package hime

import (
	"context"
	"crypto/tls"
	"html/template"
	"net"
	"net/http"
	"sync"

	"github.com/moonrhythm/parapet"
)

// App is the hime app
type App struct {
	srv           *parapet.Server
	handler       http.Handler
	routes        Routes
	globals       sync.Map
	onceServeHTTP sync.Once
	serveHandler  http.Handler

	template        map[string]*tmpl
	component       map[string]*tmpl
	cachedComponent sync.Map
	parent          *template.Template

	ETag bool
}

type ctxKeyApp struct{}

// New creates new app
func New() *App {
	app := &App{}
	app.SetServer(&parapet.Server{})
	app.setupParent()
	return app
}

// Clone clones app
func (app *App) Clone() *App {
	x := &App{
		srv: &parapet.Server{
			Addr:               app.srv.Addr,
			ReadTimeout:        app.srv.ReadTimeout,
			ReadHeaderTimeout:  app.srv.ReadHeaderTimeout,
			WriteTimeout:       app.srv.WriteTimeout,
			IdleTimeout:        app.srv.IdleTimeout,
			MaxHeaderBytes:     app.srv.MaxHeaderBytes,
			TCPKeepAlivePeriod: app.srv.TCPKeepAlivePeriod,
			GraceTimeout:       app.srv.GraceTimeout,
			WaitBeforeShutdown: app.srv.WaitBeforeShutdown,
			ErrorLog:           app.srv.ErrorLog,
			TrustProxy:         app.srv.TrustProxy,
			H2C:                app.srv.H2C,
			ReusePort:          app.srv.ReusePort,
			ConnState:          app.srv.ConnState,
			TLSConfig:          app.srv.TLSConfig.Clone(),
			BaseContext:        app.srv.BaseContext,
		},
		handler:  app.handler,
		routes:   cloneRoutes(app.routes),
		globals:  cloneMap(&app.globals),
		template: cloneTmpl(app.template),
		parent:   template.Must(app.parent.Clone()),
		ETag:     app.ETag,
	}
	x.srv.Handler = x
	x.setupParent()

	return x
}

// Address sets server address
func (app *App) Address(addr string) {
	app.srv.Addr = addr
}

// Handler sets the handler
func (app *App) Handler(h http.Handler) {
	app.handler = h
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.onceServeHTTP.Do(func() {
		app.serveHandler = app.handler
		if app.serveHandler == nil {
			app.serveHandler = http.DefaultServeMux
		}

		app.serveHandler = app.ServeHandler(app.serveHandler)
	})

	app.serveHandler.ServeHTTP(w, r)
}

func (app *App) ServeHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxKeyApp{}, app)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}

// Server returns server inside app
func (app *App) Server() *parapet.Server {
	return app.srv
}

func (app *App) SetServer(srv *parapet.Server) {
	if srv == nil {
		panic("nil server")
	}
	app.srv = srv
	srv.Handler = app
}

// Shutdown shutdowns server
func (app *App) Shutdown() error {
	return app.srv.Shutdown()
}

// ListenAndServe starts web server
func (app *App) ListenAndServe() error {
	return app.srv.ListenAndServe()
}

// Serve serves listener
func (app *App) Serve(l net.Listener) error {
	return app.srv.Serve(l)
}

func (app *App) ensureTLSConfig() {
	if app.srv.TLSConfig == nil {
		app.srv.TLSConfig = &tls.Config{}
	}
}

// TLS sets cert and key file
func (app *App) TLS(certFile, keyFile string) {
	app.ensureTLSConfig()

	err := loadTLSCertKey(app.srv.TLSConfig, certFile, keyFile)
	if err != nil {
		panicf("load key pair; %v", err)
	}
}

// SelfSign generates self sign cert
func (app *App) SelfSign(s SelfSign) {
	app.ensureTLSConfig()

	err := s.config(app.srv.TLSConfig)
	if err != nil {
		panicf("generate self sign; %v", err)
	}
}

func (app *App) setupParent() {
	if app.parent == nil {
		app.parent = template.New("")
	}
	if app.template == nil {
		app.template = make(map[string]*tmpl)
	}
	if app.component == nil {
		app.component = make(map[string]*tmpl)
	}
	app.parent.Funcs(template.FuncMap{
		"param":        tfParam,
		"templateName": func() string { return "" },
		"component":    app.renderComponent,
		"route":        app.Route,
		"global":       app.Global,
	})
}

func getApp(ctx context.Context) *App {
	app, ok := ctx.Value(ctxKeyApp{}).(*App)
	if !ok {
		panic(ErrAppNotFound)
	}
	return app
}
