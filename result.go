package hime

import (
	"bytes"
	"log"
	"net/http"

	"github.com/acoshift/header"
)

func (ctx *appContext) Redirect(url string) Result {
	if ctx.r.Method == http.MethodPost {
		return ctx.RedirectWithCode(url, http.StatusSeeOther)
	}
	return ctx.RedirectWithCode(url, http.StatusFound)
}

func (ctx *appContext) RedirectWithCode(url string, code int) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(ctx.w, ctx.r, url, code)
	})
}

func (ctx *appContext) RedirectTo(name string) Result {
	path, ok := ctx.app.namedPath[name]
	if !ok {
		log.Panicf("hime: path name %s not found", name)
	}
	return ctx.Redirect(path)
}

func (ctx *appContext) RedirectToWithCode(name string, code int) Result {
	path, ok := ctx.app.namedPath[name]
	if !ok {
		log.Panicf("hime: path name %s not found", name)
	}
	return ctx.RedirectWithCode(path, code)
}

func (ctx *appContext) Error(error string, code int) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(ctx.w, error, code)
	})
}

func (ctx *appContext) View(name string, data interface{}) Result {
	return ctx.ViewWithCode(name, http.StatusOK, data)
}

func (ctx *appContext) ViewWithCode(name string, code int, data interface{}) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}
