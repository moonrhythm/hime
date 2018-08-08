package hime

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func invokeHandler(h http.Handler, method string, target string, body io.Reader) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestResult(t *testing.T) {
	t.Parallel()

	t.Run("StatusCode", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.Status(http.StatusNotFound).String("not found")
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Equal(t, "not found", w.Body.String())
	})

	t.Run("StatusTest", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.Status(http.StatusTeapot).StatusText()
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusTeapot, w.Result().StatusCode)
		assert.Equal(t, http.StatusText(http.StatusTeapot), w.Body.String())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.NotFound()
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Equal(t, "404 page not found\n", w.Body.String())
	})

	t.Run("NoContent", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.NoContent()
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
	})

	t.Run("Bytes", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.Bytes([]byte("hello hime"))
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "hello hime", w.Body.String())
	})

	t.Run("File", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.File("testdata/file.txt")
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "file content", w.Body.String())
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.JSON(map[string]interface{}{"abc": "afg", "bbb": 123})
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"abc":"afg","bbb":123}`, w.Body.String())
	})

	t.Run("HTML", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.HTML([]byte(`<h1>Hello</h1>`))
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, `<h1>Hello</h1>`, w.Body.String())
	})

	t.Run("String", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.String("hello, hime")
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "hello, hime", w.Body.String())
	})

	t.Run("Nil", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return nil
			}))

		assert.NotPanics(t, func() {
			w := invokeHandler(app, "GET", "/", nil)
			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			assert.Empty(t, w.Body.String())
		})
	})

	t.Run("Handle", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("handler"))
				}))
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "handler", w.Body.String())
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.Error("some error :P")
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		assert.Equal(t, "some error :P\n", w.Body.String())
	})

	t.Run("ErrorCustomStatusCode", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.Status(http.StatusNotFound).Error("some not found error :P")
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		assert.Equal(t, "some not found error :P\n", w.Body.String())
	})

	t.Run("RedirectTo", func(t *testing.T) {
		t.Parallel()

		app := New().
			Routes(Routes{
				"route1": "/route/1",
			}).
			Handler(Handler(func(ctx *Context) error {
				return ctx.RedirectTo("route1")
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusFound, w.Result().StatusCode)
		l, err := w.Result().Location()
		assert.NoError(t, err)
		assert.Equal(t, "/route/1", l.String())
	})

	t.Run("RedirectToUnknownRoute", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.RedirectTo("unknown")
			}))

		assert.Panics(t, func() { invokeHandler(app, "GET", "/", nil) })
	})

	t.Run("View", func(t *testing.T) {
		t.Parallel()

		app := New()

		app.Template().Dir("testdata").Root("root").Parse("index", "hello.tmpl")

		app.
			Handler(Handler(func(ctx *Context) error {
				return ctx.View("index", nil)
			}))

		w := invokeHandler(app, "GET", "/", nil)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, "hello", w.Body.String())
	})

	t.Run("ViewNotFound", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				return ctx.View("index", nil)
			}))

		assert.Panics(t, func() { invokeHandler(app, "GET", "/", nil) })
	})

	t.Run("ViewWrongTemplateFunc", func(t *testing.T) {
		t.Parallel()

		app := New()

		app.TemplateFunc("fn", func(s string) string { return s })

		app.Template().Dir("testdata").Root("root").Parse("index", "call_fn.tmpl")

		app.Handler(Handler(func(ctx *Context) error {
			return ctx.View("index", nil)
		}))

		assert.Panics(t, func() { invokeHandler(app, "GET", "/", nil) })
	})
}

func panicRecovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func TestPanicInView(t *testing.T) {
	t.Parallel()

	t.Run("MinifyDisabled", func(t *testing.T) {
		t.Parallel()

		app := New()

		app.TemplateFunc("panic", func() string { panic("panic") })
		app.Template().Dir("testdata").Root("root").Parse("index", "panic.tmpl")

		app.
			Handler(Handler(func(ctx *Context) error {
				return ctx.View("index", nil)
			}))

		ts := httptest.NewServer(panicRecovery(app))
		defer ts.Close()

		resp, err := http.Get(ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("MinifyEnabled", func(t *testing.T) {
		t.Parallel()

		app := New()

		app.TemplateFunc("panic", func() string { panic("panic") })
		app.Template().Dir("testdata").Root("root").Parse("index", "panic.tmpl").Minify()

		app.
			Handler(Handler(func(ctx *Context) error {
				return ctx.View("index", nil)
			}))

		ts := httptest.NewServer(panicRecovery(app))
		defer ts.Close()

		resp, err := http.Get(ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
