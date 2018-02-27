package hime

import (
	"context"
	"net/http"
)

// NewContext creates new hime's context
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	app, ok := r.Context().Value(ctxKeyApp).(*App)
	if !ok {
		panic(ErrAppNotFound)
	}
	return newContext(app, w, r)
}

func newContext(app *App, w http.ResponseWriter, r *http.Request) *Context {
	return &Context{r.Context(), app, r, w, 0}
}

// Context is the hime context
type Context struct {
	context.Context

	app *App
	r   *http.Request
	w   http.ResponseWriter

	code int
}

// Request returns http.Request from context
func (ctx *Context) Request() *http.Request {
	return ctx.r
}

// ResponseWriter returns http.ResponseWriter from context
func (ctx *Context) ResponseWriter() http.ResponseWriter {
	return ctx.w
}

// Status sets response status
func (ctx *Context) Status(code int) *Context {
	ctx.code = code
	return ctx
}

// Param creates new param for redirect
func (ctx *Context) Param(name string, value interface{}) *Param {
	return &Param{Name: name, Value: value}
}
