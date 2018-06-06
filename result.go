package hime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
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

// Redirect redircets to given url
func (ctx *Context) Redirect(url string, params ...interface{}) Result {
	p := buildPath(url, params...)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(ctx.w, ctx.r, p, ctx.statusCodeRedirect())
	})
}

// SafeRedirect extracts only path from url then redirect
func (ctx *Context) SafeRedirect(url string, params ...interface{}) Result {
	p := buildPath(url, params...)
	return ctx.Redirect(SafeRedirectPath(p))
}

// RedirectTo redirects to route name
func (ctx *Context) RedirectTo(name string, params ...interface{}) Result {
	p := buildPath(ctx.app.Route(name), params...)
	return ctx.Redirect(p)
}

// RedirectToGet redirects to same url with status SeeOther
func (ctx *Context) RedirectToGet() Result {
	return ctx.Status(http.StatusSeeOther).Redirect(ctx.Request().RequestURI)
}

// RedirectBack redirects to referer or fallback if referer not exists
func (ctx *Context) RedirectBack(fallback string) Result {
	u := ctx.r.Referer()
	if u == "" {
		u = fallback
	}
	if u == "" {
		u = ctx.Request().RequestURI
	}
	return ctx.Redirect(u)
}

// RedirectBackToGet redirects to referer with status SeeOther or fallback
// with same url
func (ctx *Context) RedirectBackToGet() Result {
	return ctx.Status(http.StatusSeeOther).RedirectBack("")
}

// SafeRedirectBack safe redirects to referer
func (ctx *Context) SafeRedirectBack(fallback string) Result {
	u := ctx.r.Referer()
	if u == "" {
		u = fallback
	}
	if u == "" {
		u = ctx.Request().RequestURI
	}
	return ctx.SafeRedirect(u)
}

// Error calls http.Error
func (ctx *Context) Error(error string) Result {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(ctx.w, error, ctx.statusCodeError())
	})
}

// Nothing does nothing
func (ctx *Context) Nothing() Result {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do nothing
	})
}

// NotFound calls http.NotFound
func (ctx *Context) NotFound() Result {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}

// NoContent writes http.StatusNoContent into response writer
func (ctx *Context) NoContent() Result {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

// View renders view
func (ctx *Context) View(name string, data interface{}) Result {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, ok := ctx.app.template[name]
		if !ok {
			panic(newErrTemplateNotFound(name))
		}

		ctx.invokeBeforeRender(func() {
			ctx.setContentType("text/html; charset=utf-8")
			ctx.w.WriteHeader(ctx.statusCode())
			panicRenderError(t.Execute(ctx.w, data))
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

// JSON encodes given data into json then writes to response writer
func (ctx *Context) JSON(data interface{}) Result {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("application/json; charset=utf-8")
			ctx.writeHeader()
			json.NewEncoder(w).Encode(data)
		})
	})
}

// String writes string into response writer
func (ctx *Context) String(format string, a ...interface{}) Result {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("text/plain; charset=utf-8")
			ctx.writeHeader()
			fmt.Fprintf(w, format, a...)
		})
	})
}

// StatusText writes status text from seted status code tnto response writer
func (ctx *Context) StatusText() Result {
	return ctx.String(http.StatusText(ctx.statusCode()))
}

// CopyFrom copies src reader into response writer
func (ctx *Context) CopyFrom(src io.Reader) Result {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.invokeBeforeRender(func() {
			ctx.setContentType("application/octet-stream")
			ctx.writeHeader()
			io.Copy(w, src)
		})
	})
}

// Bytes writes bytes into response writer
func (ctx *Context) Bytes(b []byte) Result {
	return ctx.CopyFrom(bytes.NewReader(b))
}

// File serves file using http.ServeFile
func (ctx *Context) File(name string) Result {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	})
}
