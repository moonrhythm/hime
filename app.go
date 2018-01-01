package hime

import (
	"html/template"
	"mime"
	"net/http"
	"time"

	"github.com/acoshift/middleware"
	"github.com/tdewolff/minify"
)

type app struct {
	handler            http.Handler
	templateFuncs      []template.FuncMap
	templateComponents []string
	templateRoot       string
	templateDir        string
	template           map[string]*template.Template
	minifier           *minify.M
	routes             Routes
	globals            Globals
	shutdownTimeout    time.Duration
	gracefulShutdown   bool
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
	app.shutdownTimeout = defShutdownTimeout
	return app
}
