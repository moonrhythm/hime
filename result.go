package hime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
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

// Handle calls h.ServeHTTP
func (ctx *Context) Handle(h http.Handler) error {
	h.ServeHTTP(ctx.w, ctx.r)
	return nil
}

// Redirect redircets to given url
func (ctx *Context) Redirect(url string, params ...interface{}) error {
	p := buildPath(url, params...)
	http.Redirect(ctx.w, ctx.r, p, ctx.statusCodeRedirect())
	return nil
}

// SafeRedirect extracts only path from url then redirect
func (ctx *Context) SafeRedirect(url string, params ...interface{}) error {
	p := buildPath(url, params...)
	return ctx.Redirect(SafeRedirectPath(p))
}

// RedirectTo redirects to route name
func (ctx *Context) RedirectTo(name string, params ...interface{}) error {
	p := buildPath(ctx.app.Route(name), params...)
	return ctx.Redirect(p)
}

// RedirectToGet redirects to same url with status SeeOther
func (ctx *Context) RedirectToGet() error {
	return ctx.Status(http.StatusSeeOther).Redirect(ctx.Request().RequestURI)
}

// RedirectBack redirects to referer or fallback if referer not exists
func (ctx *Context) RedirectBack(fallback string) error {
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
func (ctx *Context) RedirectBackToGet() error {
	return ctx.Status(http.StatusSeeOther).RedirectBack("")
}

// SafeRedirectBack safe redirects to referer
func (ctx *Context) SafeRedirectBack(fallback string) error {
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
func (ctx *Context) Error(error string) error {
	http.Error(ctx.w, error, ctx.statusCodeError())
	return nil
}

// NotFound calls http.NotFound
func (ctx *Context) NotFound() error {
	http.NotFound(ctx.w, ctx.r)
	return nil
}

// NoContent writes http.StatusNoContent into response writer
func (ctx *Context) NoContent() error {
	ctx.w.WriteHeader(http.StatusNoContent)
	return nil
}

// View renders view
func (ctx *Context) View(name string, data interface{}) error {
	t, ok := ctx.app.template[name]
	if !ok {
		panic(newErrTemplateNotFound(name))
	}

	ctx.invokeBeforeRender(func() {
		ctx.setContentType("text/html; charset=utf-8")
		ctx.w.WriteHeader(ctx.statusCode())
		panicRenderError(t.Execute(ctx.w, data))
	})
	return nil
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
	if len(ctx.w.Header().Get("Content-Type")) == 0 {
		ctx.w.Header().Set("Content-Type", value)
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
func (ctx *Context) JSON(data interface{}) error {
	ctx.invokeBeforeRender(func() {
		ctx.setContentType("application/json; charset=utf-8")
		ctx.writeHeader()
		json.NewEncoder(ctx.w).Encode(data)
	})
	return nil
}

// String writes string into response writer
func (ctx *Context) String(format string, a ...interface{}) error {
	ctx.invokeBeforeRender(func() {
		ctx.setContentType("text/plain; charset=utf-8")
		ctx.writeHeader()
		fmt.Fprintf(ctx.w, format, a...)
	})
	return nil
}

// StatusText writes status text from seted status code tnto response writer
func (ctx *Context) StatusText() error {
	return ctx.String(http.StatusText(ctx.statusCode()))
}

// CopyFrom copies src reader into response writer
func (ctx *Context) CopyFrom(src io.Reader) error {
	ctx.invokeBeforeRender(func() {
		ctx.setContentType("application/octet-stream")
		ctx.writeHeader()
		io.Copy(ctx.w, src)
	})
	return nil
}

// Bytes writes bytes into response writer
func (ctx *Context) Bytes(b []byte) error {
	return ctx.CopyFrom(bytes.NewReader(b))
}

// File serves file using http.ServeFile
func (ctx *Context) File(name string) error {
	http.ServeFile(ctx.w, ctx.r, name)
	return nil
}
