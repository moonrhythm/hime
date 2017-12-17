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
	TemplateDir(path string) App
	TemplateRoot(name string) App
	TemplateFuncs(funcs ...template.FuncMap) App
	Component(filename ...string) App
	Template(name string, filename ...string) App
	BeforeRender(m middleware.Middleware) App
	Minify() App
	Handler(factory HandlerFactory) App
	Route(name, path string) App
	GracefulShutdown() App
	ShutdownTimeout(d time.Duration) App
	ListenAndServe(addr string) error

	GetRoute(name string) string
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
