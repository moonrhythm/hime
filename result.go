package hime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"syscall"

	"github.com/acoshift/header"
)

func (ctx *appContext) statusCode() int {
	if ctx.code == 0 {
		return http.StatusOK
	}
	return ctx.code
}

func (ctx *appContext) statusCodeRedirect() int {
	if ctx.code == 0 {
		if ctx.r.Method == http.MethodPost {
			return http.StatusSeeOther
		}
		return http.StatusFound
	}
	return ctx.code
}

func (ctx *appContext) statusCodeError() int {
	if ctx.code == 0 {
		return http.StatusInternalServerError
	}
	return ctx.code
}

func (ctx *appContext) Redirect(url string) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(ctx.w, ctx.r, url, ctx.statusCodeRedirect())
	})
}

func safeRedirectPath(p string) string {
	l, err := url.ParseRequestURI(p)
	if err != nil || len(l.Path) == 0 {
		return "/"
	}
	return l.Path
}

func (ctx *appContext) SafeRedirect(url string) Result {
	return ctx.Redirect(safeRedirectPath(url))
}

func (ctx *appContext) RedirectTo(name string) Result {
	path := ctx.app.Route(name)
	return ctx.Redirect(path)
}

func (ctx *appContext) Error(error string) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(ctx.w, error, ctx.statusCodeError())
	})
}

func (ctx *appContext) View(name string, data interface{}) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		t, ok := ctx.app.template[name]
		if !ok {
			log.Panicf("hime: template %s not found", name)
		}

		ctx.invokeBeforeRender(func() {
			ctx.renderView(t, ctx.statusCode(), data)
		})
	})
}

func (ctx *appContext) invokeBeforeRender(after func()) {
	if ctx.app.beforeRender != nil {
		ctx.app.beforeRender(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			after()
		})).ServeHTTP(ctx.w, ctx.r)
		return
	}
	after()
}

func (ctx *appContext) setContentType(value string) {
	if len(ctx.w.Header().Get(header.ContentType)) == 0 {
		ctx.w.Header().Set(header.ContentType, value)
	}
}

func panicRenderError(err error) {
	if err == nil {
		return
	}
	if _, ok := err.(*net.OpError); ok {
		return
	}
	if err == syscall.EPIPE {
		return
	}
	panic(err)
}

func (ctx *appContext) renderView(t *template.Template, code int, data interface{}) {
	ctx.setContentType("text/html; charset=utf-8")
	ctx.w.WriteHeader(code)

	if ctx.app.minifier == nil {
		err := t.Execute(ctx.w, data)
		panicRenderError(err)
		return
	}

	buf := bytes.Buffer{}
	err := t.Execute(&buf, data)
	panicRenderError(err)
	err = ctx.app.minifier.Minify("text/html", ctx.w, &buf)
	panicRenderError(err)
}

func (ctx *appContext) JSON(data interface{}) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("application/json; charset=utf-8")
			json.NewEncoder(w).Encode(data)
		})
	})
}

func (ctx *appContext) String(format string, a ...interface{}) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("text/plain; charset=utf-8")
			fmt.Fprintf(w, format, a)
		})
	})
}

func (ctx *appContext) CopyFrom(src io.Reader) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("application/octet-stream")
			io.Copy(w, src)
		})
	})
}

func (ctx *appContext) Bytes(b []byte) Result {
	return ctx.CopyFrom(bytes.NewReader(b))
}

func (ctx *appContext) Handle(h http.Handler) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(ctx.w, ctx.r)
	})
}
