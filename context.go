package hime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"
)

// NewContext creates new hime's context
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	app, ok := r.Context().Value(ctxKeyApp).(*App)
	if !ok {
		panic(ErrAppNotFound)
	}
	return NewAppContext(app, w, r)
}

// NewAppContext creates new hime's context with given app
func NewAppContext(app *App, w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request: r,
		app:     app,
		w:       w,
	}
}

// Context is hime context
type Context struct {
	*http.Request

	app *App
	w   http.ResponseWriter

	code int
}

// Deadline implements context.Context
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return ctx.Request.Context().Deadline()
}

// Done implements context.Context
func (ctx *Context) Done() <-chan struct{} {
	return ctx.Request.Context().Done()
}

// Err implements context.Context
func (ctx *Context) Err() error {
	return ctx.Request.Context().Err()
}

// Value implements context.Context
func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.Request.Context().Value(key)
}

// WithContext sets r to r.WithContext with given context
func (ctx *Context) WithContext(nctx context.Context) {
	ctx.Request = ctx.Request.WithContext(nctx)
}

// WithResponseWriter overrides response writer
func (ctx *Context) WithResponseWriter(w http.ResponseWriter) {
	ctx.w = w
}

// WithValue calls WithContext with value context
func (ctx *Context) WithValue(key interface{}, val interface{}) {
	ctx.WithContext(context.WithValue(ctx.Context(), key, val))
}

// Status sets response status code
func (ctx *Context) Status(code int) *Context {
	ctx.code = code
	return ctx
}

// Param is the short-hand for hime.Param
func (ctx *Context) Param(name string, value interface{}) *Param {
	return &Param{Name: name, Value: value}
}

func (ctx *Context) statusCode() int {
	if ctx.code == 0 {
		return http.StatusOK
	}
	return ctx.code
}

func (ctx *Context) statusCodeRedirect() int {
	if ctx.code == 0 || ctx.code < 300 || ctx.code >= 400 {
		if ctx.Request.Method == http.MethodPost {
			return http.StatusSeeOther
		}
		return http.StatusFound
	}
	return ctx.code
}

func (ctx *Context) statusCodeError() int {
	if ctx.code < 400 {
		return http.StatusInternalServerError
	}
	return ctx.code
}

func (ctx *Context) writeHeader() {
	ctx.w.WriteHeader(ctx.statusCode())
}

// Handle calls h.ServeHTTP
func (ctx *Context) Handle(h http.Handler) error {
	h.ServeHTTP(ctx.w, ctx.Request)
	return nil
}

// Redirect redircets to given url
func (ctx *Context) Redirect(url string, params ...interface{}) error {
	p := buildPath(url, params...)
	http.Redirect(ctx.w, ctx.Request, p, ctx.statusCodeRedirect())
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
	return ctx.Status(http.StatusSeeOther).Redirect(ctx.RequestURI)
}

// RedirectBack redirects to referer or fallback if referer not exists
func (ctx *Context) RedirectBack(fallback string) error {
	u := ctx.Referer()
	if u == "" {
		u = fallback
	}
	if u == "" {
		u = ctx.RequestURI
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
	u := ctx.Request.Referer()
	if u == "" {
		u = fallback
	}
	if u == "" {
		u = ctx.RequestURI
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
	http.NotFound(ctx.w, ctx.Request)
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

	buf := bytesPool.Get().(*bytes.Buffer)
	defer bytesPool.Put(buf)

	buf.Reset()
	err := t.Execute(buf, data)
	if err != nil {
		return err
	}

	return ctx.HTML(buf.Bytes())
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
	http.ServeFile(ctx.w, ctx.Request, name)
	return nil
}

// ResponseWriter returns response writer
func (ctx *Context) ResponseWriter() http.ResponseWriter {
	return ctx.w
}

// AddHeader adds a header to response
func (ctx *Context) AddHeader(key, value string) {
	ctx.w.Header().Add(key, value)
}

// AddHeaderIfNotExists adds a header to response if not exists
func (ctx *Context) AddHeaderIfNotExists(key, value string) {
	if v := ctx.w.Header().Get(key); v == "" {
		ctx.w.Header().Add(key, value)
	}
}

// SetHeader sets a header to response
func (ctx *Context) SetHeader(key, value string) {
	ctx.w.Header().Set(key, value)
}

// DelHeader deletes a header from response
func (ctx *Context) DelHeader(key string) {
	ctx.w.Header().Del(key)
}
