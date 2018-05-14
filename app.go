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

type app struct {
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
	defTemplateRoot    = "layout"
	defTemplateDir     = "view"
	defShutdownTimeout = 30 * time.Second
)

var (
	ctxKeyApp = struct{}{}
)

func init() {
	mime.AddExtensionType(".js", "text/javascript")
}

// New creates new app
func New() App {
	app := &app{}
	app.template = make(map[string]*template.Template)
	app.templateRoot = defTemplateRoot
	app.templateDir = defTemplateDir
	app.routes = make(Routes)
	app.globals = make(Globals)
	return app
}

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
func (app *app) Handler(h http.Handler) App {
	app.handler = h
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

func (app *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, ctxKeyApp, app)
	r = r.WithContext(ctx)
	app.handler.ServeHTTP(w, r)
}

func (app *app) Server(server *http.Server) App {
	app.srv = server
	return app
}

// ListenAndServe is the shotcut for http.ListenAndServe
func (app *app) ListenAndServe(addr string) error {
	if app.srv == nil {
		app.srv = &http.Server{
			Addr:    addr,
			Handler: app,
		}
	}

	return app.srv.ListenAndServe()
}

// GracefulShutdown change app to graceful mode
func (app *app) GracefulShutdown() GracefulShutdownApp {
	return &gracefulShutdownApp{
		app:     app,
		timeout: defShutdownTimeout,
	}
}
