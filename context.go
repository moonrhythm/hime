package hime

import (
	"context"
	"net/http"
)

// NewContext creates new hime's context
func NewContext(w http.ResponseWriter, r *http.Request) Context {
	return newInternalContext(w, r)
}

func newInternalContext(w http.ResponseWriter, r *http.Request) *appContext {
	app, ok := r.Context().Value(ctxKeyApp).(*app)
	if !ok {
		panic(ErrAppNotFound)
	}
	return newContext(app, w, r)
}

func newContext(app *app, w http.ResponseWriter, r *http.Request) *appContext {
	return &appContext{r.Context(), app, r, w, 0}
}

type appContext struct {
	context.Context

	app *app
	r   *http.Request
	w   http.ResponseWriter

	code int
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
