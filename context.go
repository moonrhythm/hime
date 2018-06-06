package hime

import (
	"context"
	"net/http"
	"time"
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
	return &Context{app, r, w, 0}
}

// Context is hime context
type Context struct {
	app *App
	r   *http.Request
	w   http.ResponseWriter

	code int
}

// Deadline implements context.Context
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return ctx.r.Context().Deadline()
}

// Done implements context.Context
func (ctx *Context) Done() <-chan struct{} {
	return ctx.r.Context().Done()
}

// Err implements context.Context
func (ctx *Context) Err() error {
	return ctx.r.Context().Err()
}

// Value implements context.Context
func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.r.Context().Value(key)
}

// WithContext sets r to r.WithContext with given context
func (ctx *Context) WithContext(nctx context.Context) {
	ctx.r = ctx.r.WithContext(nctx)
}

// WithRequest overrides request
func (ctx *Context) WithRequest(r *http.Request) {
	ctx.r = r
}

// WithResponseWriter overrides response writer
func (ctx *Context) WithResponseWriter(w http.ResponseWriter) {
	ctx.w = w
}

// WithValue calls WithContext with value context
func (ctx *Context) WithValue(key interface{}, val interface{}) {
	ctx.WithContext(context.WithValue(ctx.r.Context(), key, val))
}

// Request returns request
func (ctx *Context) Request() *http.Request {
	return ctx.r
}

// ResponseWriter returns response writer
func (ctx *Context) ResponseWriter() http.ResponseWriter {
	return ctx.w
}

// Status sets response status code
func (ctx *Context) Status(code int) *Context {
	ctx.code = code
	return ctx
}

// Param is the short-hand for hime.Param
func (ctx *Context) Param(name string, value interface{}) *Param {
	return &Param{Name: name, Value: value}
}
