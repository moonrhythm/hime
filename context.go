package hime

import (
	"context"
	"net/http"
)

type appContext struct {
	context.Context

	app *app
	r   *http.Request
	w   http.ResponseWriter
}

func (ctx *appContext) Request() *http.Request {
	return ctx.r
}

func (ctx *appContext) ResponseWriter() http.ResponseWriter {
	return ctx.w
}

func newContext(app *app, w http.ResponseWriter, r *http.Request) Context {
	return &appContext{r.Context(), app, r, w}
}
