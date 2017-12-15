package hime

import (
	"context"
	"html/template"
	"net/http"
	"time"
)

// App is the hime app
type App interface {
	http.Handler

	// Builder
	TemplateDir(path string) App
	TemplateRoot(name string) App
	TemplateFuncs(funcs ...template.FuncMap) App
	Component(filename ...string) App
	Template(name string, filename ...string) App
	Minify() App
	Handler(factory HandlerFactory) App
	Path(name, path string) App
	GracefulShutdown() App
	ShutdownTimeout(d time.Duration) App
	ListenAndServe(addr string) error

	GetPath(name string) string
}

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

	// Results
	Redirect(url string) Result
	RedirectWithCode(url string, code int) Result
	RedirectTo(name string) Result
	RedirectToWithCode(name string, code int) Result
	Error(error string, code int) Result
	View(name string, data interface{}) Result
	ViewWithCode(name string, code int, data interface{}) Result
	Handle(h http.Handler) Result
}
