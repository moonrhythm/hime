package hime

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"
)

// NewContext creates new hime's context
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	app := getApp(r.Context())
	return NewAppContext(app, w, r)
}

// NewAppContext creates new hime's context with given app
func NewAppContext(app *App, w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request: r,
		app:     app,
		w:       w,
		etag:    app.ETag,
	}
}

// Context is hime context
type Context struct {
	*http.Request

	app *App
	w   http.ResponseWriter

	code int
	etag bool
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
func (ctx *Context) Value(key any) any {
	return ctx.Request.Context().Value(key)
}

// WithRequest returns new context with given request
func (ctx Context) WithRequest(r *http.Request) *Context {
	ctx.Request = r
	return &ctx
}

// WithResponseWriter returns new context with given response writer
func (ctx Context) WithResponseWriter(w http.ResponseWriter) *Context {
	ctx.w = w
	return &ctx
}

// WithContext returns new context with new request with given context
func (ctx *Context) WithContext(nctx context.Context) *Context {
	return ctx.WithRequest(ctx.Request.WithContext(nctx))
}

// WithValue calls WithContext with value context
func (ctx *Context) WithValue(key any, val any) *Context {
	return ctx.WithContext(context.WithValue(ctx.Context(), key, val))
}

// Status sets response status code
func (ctx *Context) Status(code int) *Context {
	ctx.code = code
	return ctx
}

// Param is the short-hand for hime.Param
func (ctx *Context) Param(name string, value any) *Param {
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
		if ctx.Request.Method != http.MethodGet {
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

// ETag overrides etag setting
func (ctx *Context) ETag(enable bool) *Context {
	ctx.etag = enable
	return ctx
}

// Handle calls h.ServeHTTP
func (ctx *Context) Handle(h http.Handler) error {
	h.ServeHTTP(ctx.w, ctx.Request)
	return nil
}

// Redirect redirects to given url
func (ctx *Context) Redirect(url string, params ...any) error {
	p := buildPath(url, params...)
	http.Redirect(ctx.w, ctx.Request, p, ctx.statusCodeRedirect())
	return nil
}

// SafeRedirect extracts only path from url then redirect
func (ctx *Context) SafeRedirect(url string, params ...any) error {
	p := buildPath(url, params...)
	return ctx.Redirect(SafeRedirectPath(p))
}

// RedirectTo redirects to route name
func (ctx *Context) RedirectTo(name string, params ...any) error {
	p := buildPath(ctx.app.Route(name), params...)
	return ctx.Redirect(p)
}

// RedirectToGet redirects to same url back to Get
func (ctx *Context) RedirectToGet() error {
	return ctx.Redirect(ctx.RequestURI)
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

// RedirectBackToGet redirects to referer or fallback with same url
func (ctx *Context) RedirectBackToGet() error {
	return ctx.RedirectBack("")
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

func (ctx *Context) setETag(b []byte) bool {
	if ctx.etag && ctx.statusCode() == http.StatusOK {
		et := etag(b)
		ctx.w.Header().Set("ETag", et)

		if matchETag(ctx.Request, et) {
			ctx.Status(http.StatusNotModified)
			ctx.writeHeader()
			return true
		}
	}
	return false
}

// View renders view
func (ctx *Context) View(name string, data any) error {
	t, ok := ctx.app.template[name]
	if !ok {
		panic(newErrTemplateNotFound(name))
	}

	buf := getBytes()
	defer putBytes(buf)

	err := t.Execute(buf, data)
	if err != nil {
		return err
	}

	if ctx.setETag(buf.Bytes()) {
		return nil
	}

	ctx.setContentType("text/html; charset=utf-8")
	return ctx.CopyFrom(buf)
}

// Component renders component
func (ctx *Context) Component(name string, data any) error {
	t, ok := ctx.app.component[name]
	if !ok {
		panic(newErrComponentNotFound(name))
	}

	buf := getBytes()
	defer putBytes(buf)

	err := t.Execute(buf, data)
	if err != nil {
		return err
	}

	if ctx.setETag(buf.Bytes()) {
		return nil
	}

	ctx.setContentType("text/html; charset=utf-8")
	return ctx.CopyFrom(buf)
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
func (ctx *Context) JSON(data any) error {
	buf := getBytes()
	defer putBytes(buf)

	err := json.NewEncoder(buf).Encode(data)
	if err != nil {
		return err
	}

	if ctx.setETag(buf.Bytes()) {
		return nil
	}

	ctx.setContentType("application/json; charset=utf-8")
	return ctx.CopyFrom(buf)
}

// HTML writes html to response writer
func (ctx *Context) HTML(data string) error {
	if ctx.setETag([]byte(data)) {
		return nil
	}

	ctx.setContentType("text/html; charset=utf-8")
	ctx.writeHeader()
	_, err := io.Copy(ctx.w, strings.NewReader(data))
	return filterRenderError(err)
}

// String writes string into response writer
func (ctx *Context) String(format string, a ...any) error {
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

// BindJSON binds request body using json decoder
func (ctx *Context) BindJSON(v any) error {
	return json.NewDecoder(ctx.Body).Decode(v)
}

func etag(b []byte) string {
	hash := sha1.Sum(b)
	l := len(b)
	return fmt.Sprintf("W/\"%d-%s\"", l, hex.EncodeToString(hash[:]))
}

func matchETag(r *http.Request, etag string) bool {
	reqETags := strings.Split(r.Header.Get("If-None-Match"), ",")
	for _, reqETag := range reqETags {
		reqETag = strings.TrimSpace(reqETag)
		if reqETag == etag {
			return true
		}
	}
	return false
}
