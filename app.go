package hime

import (
	"context"
	"html/template"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

type app struct {
	handler            http.Handler
	templateFuncs      []template.FuncMap
	templateComponents []string
	templateRoot       string
	templateDir        string
	template           map[string]*template.Template
	minifier           *minify.M
	namedPath          map[string]string
	shutdownTimeout    time.Duration
	gracefulShutdown   bool
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
	app.namedPath = make(map[string]string)
	app.shutdownTimeout = defShutdownTimeout
	return app
}

// ShutdownTimeout sets shutdown timeout for graceful shutdown
func (app *app) ShutdownTimeout(d time.Duration) App {
	app.shutdownTimeout = d
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

// GracefulShutdown sets graceful shutdown to true
func (app *app) GracefulShutdown() App {
	app.gracefulShutdown = true
	return app
}

func (app *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, ctxKeyApp, app)
	r = r.WithContext(ctx)
	app.handler.ServeHTTP(w, r)
}

// ListenAndServe is the shotcut for http.ListenAndServe
func (app *app) ListenAndServe(addr string) (err error) {
	srv := http.Server{
		Addr:    addr,
		Handler: app,
	}

	if !app.gracefulShutdown {
		return srv.ListenAndServe()
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
		ctx, cancel := context.WithTimeout(context.Background(), app.shutdownTimeout)
		defer cancel()
		err = srv.Shutdown(ctx)
	}
	return
}
