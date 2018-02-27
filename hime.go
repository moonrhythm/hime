package hime

import (
	"net/http"
)

// Routes is the map for route name => path
type Routes map[string]string

// Globals is the global const map
type Globals map[interface{}]interface{}

// HandlerFactory is the function for create router
type HandlerFactory func(*App) http.Handler

// Factory wraps http.Handler with HandlerFactory
func Factory(h http.Handler) HandlerFactory {
	return func(_ *App) http.Handler {
		return h
	}
}

// Handler is the hime handler
type Handler func(*Context) Result

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

// Param is the query param when redirect
type Param struct {
	Name  string
	Value interface{}
}
