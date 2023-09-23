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
		tmpl := app.Template()
		tmpl.Dir("testdata")
		tmpl.Root("root")
		tmpl.ParseFiles("index", "panic.tmpl")

		app.Handler(Handler(func(ctx *Context) error {
			return ctx.View("index", nil)
		}))

		ts := httptest.NewServer(panicRecovery(app))
		defer ts.Close()

		resp, err := http.Get(ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("MinifyEnabled", func(t *testing.T) {
		t.Parallel()

		app := New()

		app.TemplateFunc("panic", func() string { panic("panic") })
		tmpl := app.Template()
		tmpl.Dir("testdata")
		tmpl.Root("root")
		tmpl.ParseFiles("index", "panic.tmpl")
		tmpl.Minify()

		app.
			Handler(Handler(func(ctx *Context) error {
				return ctx.View("index", nil)
			}))

		ts := httptest.NewServer(panicRecovery(app))
		defer ts.Close()

		resp, err := http.Get(ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusInternalServerError)
	})
}
