package hime

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

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
	_ = Context(&appContext{})
	_ = context.Context(&appContext{})
)
