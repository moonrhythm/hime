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

	buf := bytes.Buffer{}
	err := t.Execute(&buf, data)
	if err != nil {
		return err
	}

	ctx.setContentType("text/html; charset=utf-8")
	ctx.w.WriteHeader(ctx.statusCode())
	_, err = io.Copy(ctx.w, &buf)
	return filterRenderError(err)
}

func (ctx *Context) setContentType(value string) {
	if len(ctx.w.Header().Get("Content-Type")) == 0 {
		ctx.w.Header().Set("Content-Type", value)
	}
}

func filterRenderError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*net.OpError); ok {
		return nil
	}
	if err == syscall.EPIPE {
		return nil
	}
	return err
}

// JSON encodes given data into json then writes to response writer
func (ctx *Context) JSON(data interface{}) error {
	ctx.setContentType("application/json; charset=utf-8")
	ctx.writeHeader()
	return json.NewEncoder(ctx.w).Encode(data)
}

// HTML writes html to response writer
func (ctx *Context) HTML(data []byte) error {
	ctx.setContentType("text/html; charset=utf-8")
	ctx.writeHeader()
	_, err := io.Copy(ctx.w, bytes.NewReader(data))
	return filterRenderError(err)
}

// String writes string into response writer
func (ctx *Context) String(format string, a ...interface{}) error {
	ctx.setContentType("text/plain; charset=utf-8")
	ctx.writeHeader()
	_, err := fmt.Fprintf(ctx.w, format, a...)
	return filterRenderError(err)
}

// StatusText writes status text from seted status code tnto response writer
func (ctx *Context) StatusText() error {
	return ctx.String(http.StatusText(ctx.statusCode()))
}

// CopyFrom copies src reader into response writer
func (ctx *Context) CopyFrom(src io.Reader) error {
	ctx.setContentType("application/octet-stream")
	ctx.writeHeader()
	_, err := io.Copy(ctx.w, src)
	return filterRenderError(err)
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
