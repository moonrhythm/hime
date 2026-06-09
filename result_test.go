package hime

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"syscall"
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

func TestMatchETag(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("If-None-Match", `W/"1-x", W/"5-abc", W/"10-y"`)

	// matches any entry in the comma-separated list, tolerating whitespace
	assert.True(t, matchETag(r, `W/"5-abc"`))
	assert.True(t, matchETag(r, `W/"1-x"`))
	assert.False(t, matchETag(r, `W/"5-zzz"`))
}

func TestFilterRenderError(t *testing.T) {
	t.Parallel()

	// connection-style errors are swallowed
	assert.NoError(t, filterRenderError(nil))
	assert.NoError(t, filterRenderError(&net.OpError{Op: "write", Err: errors.New("broken pipe")}))
	assert.NoError(t, filterRenderError(syscall.EPIPE))

	// real errors are propagated unchanged
	realErr := errors.New("real error")
	assert.Equal(t, realErr, filterRenderError(realErr))
}
