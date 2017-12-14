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

// App is the hime app
type App struct {
	router             http.Handler
	templateFuncs      []template.FuncMap
	templateComponents []string
	templateRoot       string
	templateDir        string
	template           map[string]*template.Template
	minifier           *minify.M
	namedPath          map[string]string
	shutdownTimeout    time.Duration
}

// consts
const (
	defaultTemplateRoot    = "layout"
	defaultTemplateDir     = "view"
	defaultShutdownTimeout = 30 * time.Second
)

var (
	ctxKeyApp = struct{}{}
)

func init() {
	mime.AddExtensionType(".js", "text/javascript")
}

// New creates new app
func New() *App {
	app := &App{}
	app.template = make(map[string]*template.Template)
	app.templateRoot = defaultTemplateRoot
	app.templateDir = defaultTemplateDir
	app.namedPath = make(map[string]string)
	app.shutdownTimeout = defaultShutdownTimeout
	return app
}

// ShutdownTimeout sets shutdown timeout for graceful shutdown
func (app *App) ShutdownTimeout(d time.Duration) *App {
	app.shutdownTimeout = d
	return app
}

// TemplateRoot sets template root to select when load
func (app *App) TemplateRoot(name string) *App {
	app.templateRoot = name
	return app
}

// TemplateDir sets template dir
func (app *App) TemplateDir(path string) *App {
	app.templateDir = path
	return app
}

// Router sets app router
func (app *App) Router(h http.Handler) *App {
	app.router = h
	return app
}

// Minify sets app minifier
func (app *App) Minify() *App {
	app.minifier = minify.New()
	app.minifier.AddFunc("text/html", html.Minify)
	app.minifier.AddFunc("text/css", css.Minify)
	app.minifier.AddFunc("text/javascript", js.Minify)
	return app
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, ctxKeyApp, app)
	r = r.WithContext(ctx)
	app.router.ServeHTTP(w, r)
}

// ListenAndServe is the shotcut for http.ListenAndServe
func (app *App) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, app)
}

// ListenAndServeGracefully listens to addr with graceful shutdown
func (app *App) ListenAndServeGracefully(addr string) (err error) {
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
		ctx, cancel := context.WithTimeout(context.Background(), app.shutdownTimeout)
		defer cancel()
		err = srv.Shutdown(ctx)
	}
	return
}
