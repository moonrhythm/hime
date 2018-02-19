package hime

import (
	"context"
	"net/http"
)

// NewContext creates new hime's context
func NewContext(w http.ResponseWriter, r *http.Request) Context {
	app, ok := r.Context().Value(ctxKeyApp).(*app)
	if !ok {
		panic(ErrAppNotFound)
	}
	return newContext(app, w, r)
}

type appContext struct {
	context.Context

	app *app
	r   *http.Request
	w   http.ResponseWriter

	code int
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

func newContext(app *app, w http.ResponseWriter, r *http.Request) Context {
	return &appContext{r.Context(), app, r, w, 0}
}
