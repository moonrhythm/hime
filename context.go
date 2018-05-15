package hime

import (
	"context"
	"net/http"
	"time"
)

// NewContext creates new hime's context
func NewContext(w http.ResponseWriter, r *http.Request) Context {
	return newInternalContext(w, r)
}

func newInternalContext(w http.ResponseWriter, r *http.Request) *appContext {
	app, ok := r.Context().Value(ctxKeyApp).(*App)
	if !ok {
		panic(ErrAppNotFound)
	}
	return newContext(app, w, r)
}

func newContext(app *App, w http.ResponseWriter, r *http.Request) *appContext {
	return &appContext{app, r, w, 0}
}

type appContext struct {
	app *App
	r   *http.Request
	w   http.ResponseWriter

	code int
}

// Deadline implements context.Context
func (ctx *appContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.r.Context().Deadline()
}

// Done implements context.Context
func (ctx *appContext) Done() <-chan struct{} {
	return ctx.r.Context().Done()
}

// Err implements context.Context
func (ctx *appContext) Err() error {
	return ctx.r.Context().Err()
}

// Value implements context.Context
func (ctx *appContext) Value(key interface{}) interface{} {
	return ctx.r.Context().Value(key)
}

func (ctx *appContext) WithContext(nctx context.Context) {
	ctx.r = ctx.r.WithContext(nctx)
}

func (ctx *appContext) WithRequest(r *http.Request) {
	ctx.r = r
}

func (ctx *appContext) WithResponseWriter(w http.ResponseWriter) {
	ctx.w = w
}

func (ctx *appContext) WithValue(key interface{}, val interface{}) {
	ctx.WithContext(context.WithValue(ctx.r.Context(), key, val))
}

func (ctx *appContext) Request() *http.Request {
	return ctx.r
}

func (ctx *appContext) ResponseWriter() http.ResponseWriter {
	return ctx.w
}

func (ctx *appContext) Status(code int) Context {
	ctx.code = code
	return ctx
}

func (ctx *appContext) Param(name string, value interface{}) *Param {
	return &Param{Name: name, Value: value}
}
