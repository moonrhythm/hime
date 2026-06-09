package hime_test

import (
	"bytes"
	"context"
	"encoding/xml"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/moonrhythm/hime"
)

func TestContext(t *testing.T) {
	t.Parallel()

	t.Run("panic when create context without app", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		assert.Panics(t, func() { hime.NewContext(w, r) })
	})

	t.Run("basic data", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.Equal(t, ctx.Request, r, "ctx.Request must be given request")
		assert.Equal(t, ctx.ResponseWriter(), w, "ctx.ResponseWriter() must return given response writer")
		assert.Equal(t, ctx.Param("id", 11), &hime.Param{Name: "id", Value: 11}, "ctx.Param must returns a Param")
	})

	t.Run("Value", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx = ctx.WithValue("data", "text")
		assert.Equal(t, ctx.Value("data"), "text")
	})

	t.Run("WithRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		nr := httptest.NewRequest(http.MethodPost, "/path1", nil)
		nctx := ctx.WithRequest(nr)
		assert.Equal(t, nctx.Request, nr)
		assert.Equal(t, ctx.Request, r)
	})

	t.Run("WithResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		nw := httptest.NewRecorder()
		nctx := ctx.WithResponseWriter(nw)
		assert.Equal(t, nctx.ResponseWriter(), nw)
		assert.Equal(t, ctx.ResponseWriter(), w)
	})

	t.Run("Deadline", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		nctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		ctx = ctx.WithContext(nctx)

		dt, ok := ctx.Deadline()
		ndt, nok := nctx.Deadline()
		assert.Equal(t, dt, ndt)
		assert.Equal(t, ok, nok)
		assert.Equal(t, ctx.Done(), nctx.Done())

		cancel()
		assert.Error(t, ctx.Err())
		assert.Equal(t, ctx.Err(), nctx.Err())
	})

	t.Run("Handle", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		called := false
		assert.NoError(t, ctx.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})))

		assert.True(t, called)
	})

	t.Run("AddHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx.AddHeader("Vary", "b")
		assert.Equal(t, w.Header().Get("Vary"), "b")
	})

	t.Run("AddHeaderIfNotExists", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx.AddHeaderIfNotExists("Vary", "b")
		ctx.AddHeaderIfNotExists("Vary", "c")
		assert.Len(t, w.Header()["Vary"], 1)
		assert.Equal(t, w.Header().Get("Vary"), "b")
	})

	t.Run("SetHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx.SetHeader("Vary", "b")
		ctx.SetHeader("Vary", "c")
		assert.Len(t, w.Header()["Vary"], 1)
		assert.Equal(t, w.Header().Get("Vary"), "c")
	})

	t.Run("DelHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx.SetHeader("Vary", "b")
		ctx.DelHeader("Vary")
		assert.Empty(t, w.Header().Get("Vary"))
	})

	t.Run("Status", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(401).StatusText())
		assert.Equal(t, w.Code, 401)
	})

	t.Run("StatusText", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(http.StatusTeapot).StatusText())
		assert.Equal(t, w.Code, http.StatusTeapot)
		assert.Equal(t, w.Body.String(), http.StatusText(http.StatusTeapot))
	})

	t.Run("NoContent", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.NoContent())
		assert.Equal(t, w.Code, http.StatusNoContent)
		assert.Empty(t, w.Body.String())
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.NotFound())
		assert.Equal(t, w.Code, http.StatusNotFound)
		assert.Equal(t, w.Body.String(), "404 page not found\n")
	})

	t.Run("Bytes", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Bytes([]byte("hello hime")))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/octet-stream")
		assert.Equal(t, w.Body.String(), "hello hime")
	})

	t.Run("File", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.File("testdata/file.txt"))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
		assert.Equal(t, w.Body.String(), "file content")
	})

	t.Run("JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.JSON(map[string]any{"abc": "afg", "bbb": 123}))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json; charset=utf-8")
		assert.JSONEq(t, w.Body.String(), `{"abc":"afg","bbb":123}`)
	})

	t.Run("HTML", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.HTML(`<h1>Hello</h1>`))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
		assert.Equal(t, w.Body.String(), `<h1>Hello</h1>`)
	})

	t.Run("String", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.String("hello, hime"))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
		assert.Equal(t, w.Body.String(), "hello, hime")
	})

	t.Run("Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Error("some error"))
		assert.Equal(t, w.Code, http.StatusInternalServerError)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
		assert.Equal(t, w.Body.String(), "some error\n")
	})

	t.Run("Error with status", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(http.StatusNotFound).Error("some error"))
		assert.Equal(t, w.Code, http.StatusNotFound)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
		assert.Equal(t, w.Body.String(), "some error\n")
	})

	t.Run("Redirect to external url", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Redirect("https://google.com"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "https://google.com/")
	})

	t.Run("Redirect to internal url path", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Redirect("/signin"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/signin")
	})

	t.Run("Redirect with status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(http.StatusMovedPermanently).Redirect("/signin"))
		assert.Equal(t, w.Code, http.StatusMovedPermanently)
		assert.Equal(t, w.Header().Get("Location"), "/signin")
	})

	t.Run("Redirect with PRG", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Redirect("/signin"))
		assert.Equal(t, w.Code, http.StatusSeeOther)
		assert.Equal(t, w.Header().Get("Location"), "/signin")
	})

	t.Run("Redirect with PRG and status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(http.StatusPermanentRedirect).Redirect("/signin"))
		assert.Equal(t, w.Code, http.StatusPermanentRedirect)
		assert.Equal(t, w.Header().Get("Location"), "/signin")
	})

	t.Run("RedirectTo to valid route", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectTo("route1"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/route/1")
	})

	t.Run("RedirectTo to valid route with param", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectTo("route1", ctx.Param("id", 3)))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/route/1?id=3")
	})

	t.Run("RedirectTo to valid route with additional path", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectTo("route1", "create"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/route/1/create")
	})

	t.Run("RedirectTo to valid route with additional path and param", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectTo("route1", "create", ctx.Param("id", 3)))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/route/1/create?id=3")
	})

	t.Run("RedirectTo to valid route with status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(301).RedirectTo("route1"))
		assert.Equal(t, w.Code, http.StatusMovedPermanently)
		assert.Equal(t, w.Header().Get("Location"), "/route/1")
	})

	t.Run("RedirectTo to invalid route", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.Panics(t, func() { ctx.RedirectTo("invalid") })
	})

	t.Run("RedirectToGet from get", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/signin", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectToGet())
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), r.RequestURI)
	})

	t.Run("RedirectToGet from post", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/signin", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectToGet())
		assert.Equal(t, w.Code, http.StatusSeeOther)
		assert.Equal(t, w.Header().Get("Location"), r.RequestURI)
	})

	t.Run("RedirectBack without fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectBack(""))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/path1")
	})

	t.Run("RedirectBack with fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectBack("/path2"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/path2")
	})

	t.Run("RedirectBack with referer and without fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)
		r.Header.Set("Referer", "http://localhost/path1")

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectBack(""))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "http://localhost/path1")
	})

	t.Run("RedirectBack with referer and fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)
		r.Header.Set("Referer", "http://localhost/path1")

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectBack("/path2"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "http://localhost/path1")
	})

	t.Run("SafeRedirectBack without fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.SafeRedirectBack(""))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/path1")
	})

	t.Run("SafeRedirectBack with safe fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.SafeRedirectBack("/path2"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/path2")
	})

	t.Run("SafeRedirectBack with dangerous fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.SafeRedirectBack("https://google.com/path2"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/path2")
	})

	t.Run("SafeRedirectBack with referer and without fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)
		r.Header.Set("Referer", "http://localhost/path1")

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.SafeRedirectBack(""))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/path1")
	})

	t.Run("SafeRedirectBack with referer and fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)
		r.Header.Set("Referer", "http://localhost/path1")

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.SafeRedirectBack("/path2"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/path1")
	})

	t.Run("RedirectBackToGet from get", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectBackToGet())
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/path1")
	})

	t.Run("RedirectBackToGet from get with referer", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path1", nil)
		r.Header.Set("Referer", "http://localhost/path2")

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectBackToGet())
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "http://localhost/path2")
	})

	t.Run("RedirectBackToGet from post", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path1", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectBackToGet())
		assert.Equal(t, w.Code, http.StatusSeeOther)
		assert.Equal(t, w.Header().Get("Location"), r.RequestURI)
	})

	t.Run("RedirectBackToGet from post with referer", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path1", nil)
		r.Header.Set("Referer", "http://localhost/path2")

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectBackToGet())
		assert.Equal(t, w.Code, http.StatusSeeOther)
		assert.Equal(t, w.Header().Get("Location"), "http://localhost/path2")
	})

	t.Run("SafeRedirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.SafeRedirect("https://google.com"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/")
	})

	t.Run("View with not exist template", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.Panics(t, func() { ctx.View("invalid", nil) })
	})

	t.Run("View with valid template", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		tmpl := app.Template()
		tmpl.Dir("testdata")
		tmpl.Root("root")
		tmpl.ParseFiles("index", "hello.tmpl")

		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.View("index", nil))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
		assert.Equal(t, w.Body.String(), "hello")
	})

	t.Run("View with etag", func(t *testing.T) {
		app := hime.New()
		app.ETag = true
		tmpl := app.Template()
		tmpl.Dir("testdata")
		tmpl.Root("root")
		tmpl.ParseFiles("index", "hello.tmpl")

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		ctx := hime.NewAppContext(app, w, r)
		assert.NoError(t, ctx.View("index", nil))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
		assert.Equal(t, w.Body.String(), "hello")

		etag := w.Header().Get("ETag")
		assert.True(t, strings.HasPrefix(etag, "W/\""))
		assert.True(t, strings.HasSuffix(etag, "\""))

		// second request
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("If-None-Match", etag)
		ctx = hime.NewAppContext(app, w, r)
		assert.NoError(t, ctx.View("index", nil))
		assert.Equal(t, w.Code, http.StatusNotModified)
		assert.Empty(t, w.Header().Get("Content-Type"))
		assert.Empty(t, w.Body.String())
	})

	t.Run("View with valid template and status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		tmpl := app.Template()
		tmpl.Dir("testdata")
		tmpl.Root("root")
		tmpl.ParseFiles("index", "hello.tmpl")

		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(http.StatusInternalServerError).View("index", nil))
		assert.Equal(t, w.Code, http.StatusInternalServerError)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
		assert.Equal(t, w.Body.String(), "hello")
	})

	t.Run("View with valid template and funcs invoke wrong argument", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		app.TemplateFuncs(template.FuncMap{
			"fn": func(s string) string { return s },
		})
		tmpl := app.Template()
		tmpl.Dir("testdata")
		tmpl.Root("root")
		tmpl.ParseFiles("index", "call_fn.tmpl")

		ctx := hime.NewAppContext(app, w, r)

		assert.Error(t, ctx.View("index", nil))
	})

	t.Run("View with valid template and funcs invoke panic func", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		app.TemplateFuncs(template.FuncMap{
			"panic": func() string { panic("panic") },
		})
		tmpl := app.Template()
		tmpl.Dir("testdata")
		tmpl.Root("root")
		tmpl.ParseFiles("index", "panic.tmpl")

		ctx := hime.NewAppContext(app, w, r)

		assert.Error(t, ctx.View("index", nil))
	})

	t.Run("Component not exists", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.Panics(t, func() { ctx.Component("invalid", nil) })
	})

	t.Run("Component no data", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		app.Template().Component(template.Must(template.New("c").Parse(`component`)))
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Component("c", nil))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
		assert.Equal(t, w.Body.String(), "component")
	})

	t.Run("Component with data", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		app.Template().Component(template.Must(template.New("c").Parse(`hello, {{.}}`)))
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Component("c", "world"))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
		assert.Equal(t, w.Body.String(), "hello, world")
	})

	t.Run("Render", func(t *testing.T) {
		app := hime.New()

		{
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", nil)

			ctx := hime.NewAppContext(app, w, r)

			assert.NoError(t, ctx.Render(`hello, {{.}}`, "world"))
			assert.Equal(t, w.Code, http.StatusOK)
			assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
			assert.Equal(t, w.Body.String(), "hello, world")
		}

		{
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", nil)

			ctx := hime.NewAppContext(app, w, r)

			assert.NoError(t, ctx.Render(`hello, {{.}}`, "sekai"))
			assert.Equal(t, w.Code, http.StatusOK)
			assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
			assert.Equal(t, w.Body.String(), "hello, sekai")
		}
	})

	t.Run("BindJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"a":1}`)))

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)
		var body struct {
			A int `json:"a"`
		}
		assert.NoError(t, ctx.BindJSON(&body))
		assert.Equal(t, 1, body.A)
	})

	t.Run("Cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "a", Value: "1"})

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.Equal(t, "1", ctx.CookieValue("a"))

		ctx.AddCookie("b", "2", &hime.CookieOptions{Path: "/"})
		ctx.DelCookie("a", nil)

		if !assert.Len(t, w.Header()["Set-Cookie"], 2) {
			return
		}
		assert.Equal(t, "b=2; Path=/", w.Header()["Set-Cookie"][0])
		assert.Equal(t, "a=; Max-Age=0", w.Header()["Set-Cookie"][1])
	})
}

func TestContextETagNon200(t *testing.T) {
	t.Parallel()

	// ETag is only computed for 200 responses.
	app := hime.New()
	app.ETag = true

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(app, w, r)

	assert.NoError(t, ctx.Status(http.StatusCreated).JSON(map[string]any{"a": 1}))
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Empty(t, w.Header().Get("ETag"))
}

// etag304Case runs render once to capture the ETag, then re-runs with
// If-None-Match and asserts a 304 with an empty body.
func etag304Case(t *testing.T, render func(ctx *hime.Context) error) {
	t.Helper()

	app := hime.New()
	app.ETag = true

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(app, w, r)
	assert.NoError(t, render(ctx))
	etag := w.Header().Get("ETag")
	if !assert.NotEmpty(t, etag) {
		return
	}

	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("If-None-Match", etag)
	ctx2 := hime.NewAppContext(app, w2, r2)
	assert.NoError(t, render(ctx2))
	assert.Equal(t, http.StatusNotModified, w2.Code)
	assert.Empty(t, w2.Body.String())
}

func TestContextETag304(t *testing.T) {
	t.Parallel()

	t.Run("JSON", func(t *testing.T) {
		etag304Case(t, func(ctx *hime.Context) error {
			return ctx.JSON(map[string]any{"a": 1})
		})
	})

	t.Run("HTML", func(t *testing.T) {
		etag304Case(t, func(ctx *hime.Context) error {
			return ctx.HTML(`<h1>hi</h1>`)
		})
	})

	t.Run("Render", func(t *testing.T) {
		etag304Case(t, func(ctx *hime.Context) error {
			return ctx.Render(`hello, {{.}}`, "world")
		})
	})

	t.Run("XML", func(t *testing.T) {
		etag304Case(t, func(ctx *hime.Context) error {
			return ctx.XML(struct {
				XMLName xml.Name `xml:"item"`
				V       string   `xml:"v"`
			}{V: "x"})
		})
	})
}

func TestContextComponentETag304(t *testing.T) {
	t.Parallel()

	app := hime.New()
	app.ETag = true
	app.Template().Component(template.Must(template.New("c").Parse(`component`)))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(app, w, r)
	assert.NoError(t, ctx.Component("c", nil))
	etag := w.Header().Get("ETag")
	if !assert.NotEmpty(t, etag) {
		return
	}

	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("If-None-Match", etag)
	ctx2 := hime.NewAppContext(app, w2, r2)
	assert.NoError(t, ctx2.Component("c", nil))
	assert.Equal(t, http.StatusNotModified, w2.Code)
	assert.Empty(t, w2.Body.String())
}

func TestContextDelCookieWithOptions(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	ctx.DelCookie("session", &hime.CookieOptions{
		Path:     "/admin",
		Domain:   "example.com",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	sc := w.Header().Get("Set-Cookie")
	assert.Contains(t, sc, "session=")
	assert.Contains(t, sc, "Max-Age=0")
	assert.Contains(t, sc, "Path=/admin")
	assert.Contains(t, sc, "Domain=example.com")
	assert.Contains(t, sc, "Secure")
	assert.Contains(t, sc, "HttpOnly")
	assert.Contains(t, sc, "SameSite=Lax")
}

func TestContextAddCookieNilOptions(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	ctx.AddCookie("x", "y", nil)
	assert.Equal(t, "x=y", w.Header().Get("Set-Cookie"))
}

func TestContextCookieValueMissing(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.Equal(t, "", ctx.CookieValue("missing"))
}

func TestContextJSONError(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.Error(t, ctx.JSON(make(chan int)))
}

func TestContextBindJSONInvalid(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{invalid"))
	ctx := hime.NewAppContext(hime.New(), w, r)

	var v map[string]any
	assert.Error(t, ctx.BindJSON(&v))
}

func TestContextFileNotFound(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.NoError(t, ctx.File("testdata/does-not-exist.txt"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestContextRedirectInvalidURL(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.Panics(t, func() { ctx.Redirect("%zz") })
}

func TestContextSafeRedirectDangerous(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.NoError(t, ctx.SafeRedirect("https://evil.com/path"))
	assert.Equal(t, "/path", w.Header().Get("Location"))
}

func TestContextStringFormat(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.NoError(t, ctx.String("hello %s %d", "world", 42))
	assert.Equal(t, "hello world 42", w.Body.String())
}

func TestContextSetContentTypeRespected(t *testing.T) {
	t.Parallel()

	// a pre-set Content-Type is not overridden by a write method
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	ctx.SetHeader("Content-Type", "application/xml")
	assert.NoError(t, ctx.HTML("<x/>"))
	assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))
}

func TestContextErrorStatusCode(t *testing.T) {
	t.Parallel()

	t.Run("preserves 4xx", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := hime.NewAppContext(hime.New(), w, r)

		assert.NoError(t, ctx.Status(http.StatusUnauthorized).Error("nope"))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("promotes <400 to 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := hime.NewAppContext(hime.New(), w, r)

		assert.NoError(t, ctx.Status(http.StatusOK).Error("nope"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestContextRenderParseError(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.Error(t, ctx.Render("{{.", nil))
}

func TestContextXML(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	payload := struct {
		XMLName xml.Name `xml:"item"`
		Name    string   `xml:"name"`
		Value   int      `xml:"value"`
	}{Name: "hime", Value: 7}

	assert.NoError(t, ctx.XML(payload))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/xml; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, "<item><name>hime</name><value>7</value></item>", w.Body.String())
}

func TestContextXMLError(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	// channels cannot be marshalled to xml
	assert.Error(t, ctx.XML(make(chan int)))
}

func TestContextBindXML(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`<item><name>hime</name></item>`))
	ctx := hime.NewAppContext(hime.New(), w, r)

	var body struct {
		Name string `xml:"name"`
	}
	assert.NoError(t, ctx.BindXML(&body))
	assert.Equal(t, "hime", body.Name)
}

func TestContextBindXMLInvalid(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`<item><name>hime`))
	ctx := hime.NewAppContext(hime.New(), w, r)

	var body struct {
		Name string `xml:"name"`
	}
	assert.Error(t, ctx.BindXML(&body))
}

func TestContextStatusCode(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.Equal(t, http.StatusOK, ctx.StatusCode()) // default when unset
	ctx.Status(http.StatusTeapot)
	assert.Equal(t, http.StatusTeapot, ctx.StatusCode())
}

func TestContextRenderComponentToString(t *testing.T) {
	t.Parallel()

	app := hime.New()
	app.Template().Component(template.Must(template.New("badge").Parse(`<b>{{.}}</b>`)))
	ctx := hime.NewAppContext(app, httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	s, err := ctx.RenderComponentToString("badge", "new")
	assert.NoError(t, err)
	assert.Equal(t, "<b>new</b>", s)

	assert.Panics(t, func() { ctx.RenderComponentToString("missing", nil) })

	// execute errors are returned
	app.Template().Component(template.Must(template.New("bad").Parse(`{{.A.B}}`)))
	_, err = ctx.RenderComponentToString("bad", map[string]any{"A": "not-a-struct"})
	assert.Error(t, err)
}

func TestContextETagOverride(t *testing.T) {
	t.Parallel()

	t.Run("enable on a non-ETag app", func(t *testing.T) {
		app := hime.New() // app.ETag defaults to false
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.ETag(true).JSON(map[string]any{"a": 1}))
		assert.NotEmpty(t, w.Header().Get("ETag"))
	})

	t.Run("disable on an ETag app", func(t *testing.T) {
		app := hime.New()
		app.ETag = true
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.ETag(false).JSON(map[string]any{"a": 1}))
		assert.Empty(t, w.Header().Get("ETag"))
	})
}
