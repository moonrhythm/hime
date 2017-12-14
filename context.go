package hime

import (
	"bytes"
	"context"
	"log"
	"net/http"

	"github.com/acoshift/header"
)

// Context is the hime context
type Context interface {
	context.Context

	Request() *http.Request
	ResponseWriter() http.ResponseWriter

	Redirect(url string)
	RedirectWithCode(url string, code int)
	RedirectTo(name string)
	RedirectToWithCode(name string, code int)

	Error(error string, code int)

	View(name string, data interface{})
	ViewWithCode(name string, code int, data interface{})
}

type appContext struct {
	context.Context

	app *App
	r   *http.Request
	w   http.ResponseWriter
}

func (ctx *appContext) Request() *http.Request {
	return ctx.r
}

func (ctx *appContext) ResponseWriter() http.ResponseWriter {
	return ctx.w
}

func (ctx *appContext) Redirect(url string) {
	if ctx.r.Method == http.MethodPost {
		ctx.RedirectWithCode(url, http.StatusSeeOther)
		return
	}
	ctx.RedirectWithCode(url, http.StatusFound)
}

func (ctx *appContext) RedirectWithCode(url string, code int) {
	http.Redirect(ctx.w, ctx.r, url, code)
}

func (ctx *appContext) RedirectTo(name string) {
	path, ok := ctx.app.namedPath[name]
	if !ok {
		log.Panicf("hime: path name %s not found", name)
	}
	ctx.Redirect(path)
}

func (ctx *appContext) RedirectToWithCode(name string, code int) {
	path, ok := ctx.app.namedPath[name]
	if !ok {
		log.Panicf("hime: path name %s not found", name)
	}
	ctx.RedirectWithCode(path, code)
}

func (ctx *appContext) Error(error string, code int) {
	http.Error(ctx.w, error, code)
}

func (ctx *appContext) View(name string, data interface{}) {
	ctx.ViewWithCode(name, http.StatusOK, data)
}

func (ctx *appContext) ViewWithCode(name string, code int, data interface{}) {
	t, ok := ctx.app.template[name]
	if !ok {
		log.Panicf("hime: template %s not found", name)
	}

	wh := ctx.w.Header()
	wh.Set(header.ContentType, "text/html; charset=utf-8")
	wh.Set(header.CacheControl, "no-cache, no-store, must-revalidate, max-age=0")
	ctx.w.WriteHeader(code)

	if ctx.app.minifier == nil {
		err := t.Execute(ctx.w, data)
		if err != nil {
			panic(err)
		}
		return
	}

	buf := bytes.Buffer{}
	err := t.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	ctx.app.minifier.Minify("text/html", ctx.w, &buf)
}

func newContext(app *App, w http.ResponseWriter, r *http.Request) Context {
	return &appContext{r.Context(), app, r, w}
}
