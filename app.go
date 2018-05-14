package hime

import (
	"context"
	"html/template"
	"mime"
	"net/http"
	"time"

	"github.com/acoshift/middleware"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

// App is the hime app
type App struct {
	srv                *http.Server
	handler            http.Handler
	templateFuncs      []template.FuncMap
	templateComponents []string
	templateRoot       string
	templateDir        string
	template           map[string]*template.Template
	minifier           *minify.M
	routes             Routes
	globals            Globals
	beforeRender       middleware.Middleware
}

// consts
const (
	defShutdownTimeout = 30 * time.Second
)

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

// Minify enables minify when render html, css, js
func (app *App) Minify() *App {
	app.minifier = minify.New()
	app.minifier.AddFunc("text/html", html.Minify)
	app.minifier.AddFunc("text/css", css.Minify)
	app.minifier.AddFunc("text/javascript", js.Minify)
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

// Server overrides server when calling ListenAndServe
func (app *App) Server(server *http.Server) *App {
	app.srv = server
	return app
}

// ListenAndServe starts web server
func (app *App) ListenAndServe(addr string) error {
	if app.srv == nil {
		app.srv = &http.Server{
			Addr:    addr,
			Handler: app,
		}
	}

	return app.srv.ListenAndServe()
}

// GracefulShutdown runs server as graceful shutdown,
// can works only when start server with app.ListenAndServe
func (app *App) GracefulShutdown() *GracefulShutdownApp {
	return &GracefulShutdownApp{
		App:     app,
		timeout: defShutdownTimeout,
	}
}
