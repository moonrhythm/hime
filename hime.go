package hime

import (
	"context"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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

	// Handler sets the handler
	Handler(h http.Handler) App

	// Routes registers route name and path
	Routes(routes Routes) App

	// Globals registers global constants
	Globals(Globals) App

	// Server overrides server when calling ListenAndServe
	Server(server *http.Server) App

	// GracefulShutdown runs server as graceful shutdown,
	// can works only when start server with app.ListenAndServe
	GracefulShutdown() GracefulShutdownApp

	// ListenAndServe starts web server
	ListenAndServe(addr string) error

	// Route gets route path from given name
	Route(name string, params ...interface{}) string

	// Global gets value from global storage
	Global(key interface{}) interface{}
}

// GracefulShutdownApp is the app in graceful shutdown mode
type GracefulShutdownApp interface {
	// Timeout sets timeout
	Timeout(d time.Duration) GracefulShutdownApp

	// Wait sets wait time before shutdown
	Wait(d time.Duration) GracefulShutdownApp

	// Notify calls fn when receive terminate signal from os
	Notify(fn func()) GracefulShutdownApp

	// Before runs fn before start waiting to SIGTERM
	Before(fn func()) GracefulShutdownApp

	// ListenAndServe starts web server
	ListenAndServe(addr string) error
}

// Routes is the map for route name => path
type Routes map[string]string

// Globals is the global const map
type Globals map[interface{}]interface{}

// Handler is the hime handler
type Handler func(Context) Result

// Result is the handler result
type Result http.Handler

// Context is the hime context
type Context interface {
	context.Context

	WithContext(ctx context.Context)
	WithRequest(r *http.Request)
	WithResponseWriter(w http.ResponseWriter)
	WithValue(key interface{}, val interface{})

	// App data

	// Route gets route path from given name
	Route(name string, params ...interface{}) string

	// Global gets value from global storage
	Global(key interface{}) interface{}

	// HTTP data

	// Request returns http.Request from context
	Request() *http.Request

	// ResponseWrite returns http.ResponseWriter from context
	ResponseWriter() http.ResponseWriter

	// Status sets response status
	Status(code int) Context

	// Request functions
	ParseForm() error
	ParseMultipartForm(maxMemory int64) error
	Form() url.Values
	PostForm() url.Values

	// FromValue functions
	FormValue(key string) string
	FormValueTrimSpace(key string) string
	FormValueTrimSpaceComma(key string) string
	FormValueInt(key string) int
	FormValueInt64(key string) int64
	FormValueFloat32(key string) float32
	FormValueFloat64(key string) float64

	PostFormValue(key string) string
	PostFormValueTrimSpace(key string) string
	PostFormValueTrimSpaceComma(key string) string
	PostFormValueInt(key string) int
	PostFormValueInt64(key string) int64
	PostFormValueFloat32(key string) float32
	PostFormValueFloat64(key string) float64

	FormFile(key string) (multipart.File, *multipart.FileHeader, error)

	// FormFileNotEmpty calls r.FormFile but return http.ErrMissingFile if file empty
	FormFileNotEmpty(key string) (multipart.File, *multipart.FileHeader, error)

	MultipartForm() *multipart.Form
	MultipartReader() (*multipart.Reader, error)
	Method() string

	// Query returns ctx.Request().URL.Query()
	Query() url.Values

	Param(name string, value interface{}) *Param

	// Results

	// Nothing does nothing
	Nothing() Result

	// Redirect redirects to given url
	Redirect(url string, params ...interface{}) Result

	// SafeRedirect extracts only path from url then redirect
	SafeRedirect(url string, params ...interface{}) Result

	// RedirectTo redirects to named route
	RedirectTo(name string, params ...interface{}) Result

	// RedirectToGet redirects to GET method with See Other status code on the current path
	RedirectToGet() Result

	// RedirectBack redirects back to previous URL
	RedirectBack(fallback string) Result

	// RedirectBackToGet redirects back to GET method with See Other status code to previous URL
	// or fallback to same URL like RedirectToGet
	RedirectBackToGet() Result

	// SafeRedirectBack redirects back to previous URL using SafeRedirect
	SafeRedirectBack(fallback string) Result

	// Error wraps http.Error
	Error(error string) Result

	// NotFound wraps http.NotFound
	NotFound() Result

	// NoContent renders empty body with http.StatusNoContent
	NoContent() Result

	// View renders template
	View(name string, data interface{}) Result

	// JSON renders json
	JSON(data interface{}) Result

	// String renders string with format
	String(format string, a ...interface{}) Result

	// StatusText renders String when http.StatusText
	StatusText() Result

	// CopyFrom copies source into response writer
	CopyFrom(src io.Reader) Result

	// Bytes renders bytes
	Bytes(b []byte) Result

	// File renders file
	File(name string) Result
}

// Param is the query param when redirect
type Param struct {
	Name  string
	Value interface{}
}

var (
	_ = App(&app{})
	_ = Context(&appContext{})
)
