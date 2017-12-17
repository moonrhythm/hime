package hime

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/acoshift/middleware"
)

// App is the hime app
type App interface {
	http.Handler

	// Builder

	// TemplateDir sets directory to load template,
	// default is "view"
	TemplateDir(path string) App

	// TemplateRoot sets root layout using t.Lookup,
	// default is "layout"
	TemplateRoot(name string) App

	// TemplateFuncs adds template funcs while load template
	TemplateFuncs(funcs ...template.FuncMap) App

	// Component adds given templates to every templates
	Component(filename ...string) App

	// Template loads template into memory
	Template(name string, filename ...string) App

	// BeforeRender runs given middleware for before render,
	// ex. View, JSON, String, Bytes, CopyForm, etc.
	BeforeRender(m middleware.Middleware) App

	// Minify enables minify when render html, css, js
	Minify() App

	// Handler sets the handler factory
	Handler(factory HandlerFactory) App

	// Routes registers route name and path
	Routes(routes Routes) App

	// GracefulShutdown runs server as graceful shutdown,
	// can works only when start server with app.ListenAndServe
	GracefulShutdown() App

	// ShutdownTimeout sets graceful shutdown timeout
	ShutdownTimeout(d time.Duration) App

	// ListenAndServe starts web server
	ListenAndServe(addr string) error

	// Route gets route path from given name
	Route(name string) string
}

// Routes is the map for route name => path
type Routes map[string]string

// HandlerFactory is the function for create router
type HandlerFactory func(App) http.Handler

// Handler is the hime handler
type Handler func(Context) Result

// Result is the handler result
type Result interface {
	Response(w http.ResponseWriter, r *http.Request)
}

// ResultFunc is the result function
type ResultFunc func(w http.ResponseWriter, r *http.Request)

// Response implements Result interface
func (f ResultFunc) Response(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

// Context is the hime context
type Context interface {
	context.Context

	Request() *http.Request
	ResponseWriter() http.ResponseWriter

	Status(code int) Context

	// Results
	Redirect(url string) Result
	SafeRedirect(url string) Result
	RedirectTo(name string) Result
	Error(error string) Result
	View(name string, data interface{}) Result
	JSON(data interface{}) Result
	String(format string, a ...interface{}) Result
	CopyFrom(src io.Reader) Result
	Bytes(b []byte) Result
	Handle(h http.Handler) Result
}
