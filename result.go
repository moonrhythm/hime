package hime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
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

func (ctx *appContext) writeHeader() {
	if code := ctx.statusCode(); code > 0 {
		ctx.w.WriteHeader(code)
	}
}

func (ctx *appContext) Redirect(url string, params ...interface{}) Result {
	p := buildPath(url, params...)
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(ctx.w, ctx.r, p, ctx.statusCodeRedirect())
	})
}

func safeRedirectPath(p string) string {
	l, err := url.ParseRequestURI(p)
	if err != nil {
		return "/"
	}
	r := l.EscapedPath()
	if len(r) == 0 {
		r = "/"
	}
	if l.ForceQuery || l.RawQuery != "" {
		r += "?" + l.RawQuery
	}
	return path.Clean(r)
}

func (ctx *appContext) SafeRedirect(url string, params ...interface{}) Result {
	p := buildPath(url, params...)
	return ctx.Redirect(safeRedirectPath(p))
}

func (ctx *appContext) RedirectTo(name string, params ...interface{}) Result {
	p := buildPath(ctx.app.Route(name), params...)
	return ctx.Redirect(p)
}

func (ctx *appContext) RedirectToGet() Result {
	return ctx.Status(http.StatusSeeOther).Redirect(ctx.Request().RequestURI)
}

func (ctx *appContext) Error(error string) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(ctx.w, error, ctx.statusCodeError())
	})
}

func (ctx *appContext) Nothing() Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		// do nothing
	})
}

func (ctx *appContext) NotFound() Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}

func (ctx *appContext) NoContent() Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

func (ctx *appContext) View(name string, data interface{}) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		t, ok := ctx.app.template[name]
		if !ok {
			panic(newErrTemplateNotFound(name))
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
			ctx.writeHeader()
			json.NewEncoder(w).Encode(data)
		})
	})
}

func (ctx *appContext) String(format string, a ...interface{}) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("text/plain; charset=utf-8")
			ctx.writeHeader()
			fmt.Fprintf(w, format, a...)
		})
	})
}

func (ctx *appContext) StatusText() Result {
	return ctx.String(http.StatusText(ctx.statusCode()))
}

func (ctx *appContext) CopyFrom(src io.Reader) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("application/octet-stream")
			ctx.writeHeader()
			io.Copy(w, src)
		})
	})
}

func (ctx *appContext) Bytes(b []byte) Result {
	return ctx.CopyFrom(bytes.NewReader(b))
}

func (ctx *appContext) File(name string) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	})
}

func (ctx *appContext) Handle(h http.Handler) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(ctx.w, ctx.r)
	})
}
