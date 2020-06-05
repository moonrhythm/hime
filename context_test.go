package hime_test

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
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

		assert.NoError(t, ctx.JSON(map[string]interface{}{"abc": "afg", "bbb": 123}))
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
		assert.Equal(t, w.Header().Get("Location"), "https://google.com")
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
		app.Template().Dir("testdata").Root("root").ParseFiles("index", "hello.tmpl")
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.View("index", nil))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
		assert.Equal(t, w.Body.String(), "hello")
	})

	t.Run("View with valid template and status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		app.Template().Dir("testdata").Root("root").ParseFiles("index", "hello.tmpl")
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
		app.Template().Dir("testdata").Root("root").ParseFiles("index", "call_fn.tmpl")
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
		app.Template().Dir("testdata").Root("root").ParseFiles("index", "panic.tmpl")
		ctx := hime.NewAppContext(app, w, r)

		assert.Error(t, ctx.View("index", nil))
	})
}
