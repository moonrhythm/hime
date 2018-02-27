package hime

import (
	"time"

	"github.com/acoshift/middleware"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

// ShutdownTimeout sets graceful shutdown timeout
func (app *App) ShutdownTimeout(d time.Duration) *App {
	app.shutdownTimeout = d
	return app
}

// TemplateRoot sets root layout using t.Lookup,
// default is "layout"
func (app *App) TemplateRoot(name string) *App {
	app.templateRoot = name
	return app
}

// TemplateDir sets directory to load template,
// default is "view"
func (app *App) TemplateDir(path string) *App {
	app.templateDir = path
	return app
}

// Handler sets the handler factory
func (app *App) Handler(factory HandlerFactory) *App {
	app.handler = factory(app)
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
// ex. View, JSON, String, Bytes, CopyForm, etc.
func (app *App) BeforeRender(m middleware.Middleware) *App {
	app.beforeRender = m
	return app
}

// GracefulShutdown runs server as graceful shutdown,
// can works only when start server with app.ListenAndServe
func (app *App) GracefulShutdown() *App {
	app.gracefulShutdown = true
	return app
}
