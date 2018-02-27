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

func (ctx *Context) statusCode() int {
	if ctx.code == 0 {
		return http.StatusOK
	}
	return ctx.code
}

func (ctx *Context) statusCodeRedirect() int {
	if ctx.code == 0 {
		if ctx.r.Method == http.MethodPost {
			return http.StatusSeeOther
		}
		return http.StatusFound
	}
	return ctx.code
}

func (ctx *Context) statusCodeError() int {
	if ctx.code == 0 {
		return http.StatusInternalServerError
	}
	return ctx.code
}

func (ctx *Context) writeHeader() {
	if code := ctx.statusCode(); code > 0 {
		ctx.w.WriteHeader(code)
	}
}

// Redirect redirects to given url
func (ctx *Context) Redirect(url string, params ...interface{}) Result {
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

// SafeRedirect extracts only path from url then redirect
func (ctx *Context) SafeRedirect(url string, params ...interface{}) Result {
	p := buildPath(url, params...)
	return ctx.Redirect(safeRedirectPath(p))
}

// RedirectTo redirects to named route
func (ctx *Context) RedirectTo(name string, params ...interface{}) Result {
	p := buildPath(ctx.app.Route(name), params...)
	return ctx.Redirect(p)
}

// RedirectToGet redirects to GET method with See Other status code on the current path
func (ctx *Context) RedirectToGet() Result {
	return ctx.Status(http.StatusSeeOther).Redirect(ctx.Request().RequestURI)
}

// Error wraps http.Error
func (ctx *Context) Error(error string) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(ctx.w, error, ctx.statusCodeError())
	})
}

// Nothing does nothing
func (ctx *Context) Nothing() Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		// do nothing
	})
}

// NotFound wraps http.NotFound
func (ctx *Context) NotFound() Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}

// NoContent renders empty body with http.StatusNoContent
func (ctx *Context) NoContent() Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

// View renders template
func (ctx *Context) View(name string, data interface{}) Result {
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

func (ctx *Context) invokeBeforeRender(after func()) {
	if ctx.app.beforeRender != nil {
		ctx.app.beforeRender(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			after()
		})).ServeHTTP(ctx.w, ctx.r)
		return
	}
	after()
}

func (ctx *Context) setContentType(value string) {
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

func (ctx *Context) renderView(t *template.Template, code int, data interface{}) {
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

// JSON renders json
func (ctx *Context) JSON(data interface{}) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("application/json; charset=utf-8")
			ctx.writeHeader()
			json.NewEncoder(w).Encode(data)
		})
	})
}

// String renders string with format
func (ctx *Context) String(format string, a ...interface{}) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("text/plain; charset=utf-8")
			ctx.writeHeader()
			fmt.Fprintf(w, format, a...)
		})
	})
}

// StatusText renders String when http.StatusText
func (ctx *Context) StatusText() Result {
	return ctx.String(http.StatusText(ctx.statusCode()))
}

// CopyFrom copies source into response writer
func (ctx *Context) CopyFrom(src io.Reader) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("application/octet-stream")
			ctx.writeHeader()
			io.Copy(w, src)
		})
	})
}

// Bytes renders bytes
func (ctx *Context) Bytes(b []byte) Result {
	return ctx.CopyFrom(bytes.NewReader(b))
}

// File renders file
func (ctx *Context) File(name string) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	})
}

// Handle wrap h with Result
func (ctx *Context) Handle(h http.Handler) Result {
	return ResultFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(ctx.w, ctx.r)
	})
}
