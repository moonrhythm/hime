package hime

import (
	"time"

	"github.com/acoshift/middleware"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

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

func (app *app) BeforeRender(m middleware.Middleware) App {
	app.beforeRender = m
	return app
}

// GracefulShutdown sets graceful shutdown to true
func (app *app) GracefulShutdown() App {
	app.gracefulShutdown = true
	return app
}
