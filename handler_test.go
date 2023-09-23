package hime

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	t.Run("panic on error", func(t *testing.T) {
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return fmt.Errorf("must panic")
		}))

		assert.Panics(t, func() {
			invokeHandler(app, "GET", "/", nil)
		})
	})

	t.Run("net/http", func(t *testing.T) {
		app := New()
		app.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}))

		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/", nil, "ok")
	})

	t.Run("hime", func(t *testing.T) {
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return ctx.String("ok")
		}))

		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/", nil, "ok")
	})

	t.Run("default handler", func(t *testing.T) {
		app := New()
		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/", nil, "404")
	})

	t.Run("cancel context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return ctx.Err()
		}))

		r := httptest.NewRequest("GET", "/", nil)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		cancel()

		assert.NotPanics(t, func() {
			app.ServeHTTP(w, r)
		})
	})
}
