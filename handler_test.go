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

	t.Run("wrapped context.Canceled does not panic", func(t *testing.T) {
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return fmt.Errorf("operation failed: %w", context.Canceled)
		}))

		assert.NotPanics(t, func() {
			invokeHandler(app, "GET", "/", nil)
		})
	})

	t.Run("context.DeadlineExceeded panics", func(t *testing.T) {
		// Only context.Canceled is swallowed; DeadlineExceeded is treated
		// as a real error and panics.
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return context.DeadlineExceeded
		}))

		assert.Panics(t, func() {
			invokeHandler(app, "GET", "/", nil)
		})
	})

	t.Run("Error does not panic", func(t *testing.T) {
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return ctx.Status(http.StatusForbidden).Error("nope")
		}))

		w := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		})
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "nope\n", w.Body.String())
	})

	t.Run("Forbidden does not panic", func(t *testing.T) {
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			if err := ctx.Forbidden("nope"); err != nil {
				return err
			}
			return ctx.String("page")
		}))

		w := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		})
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "nope\n", w.Body.String())
	})

	t.Run("NotFound does not panic", func(t *testing.T) {
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return ctx.NotFound()
		}))

		w := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		})
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "404 page not found\n", w.Body.String())
	})

	t.Run("wrapped ErrResponseWritten does not panic", func(t *testing.T) {
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return fmt.Errorf("load: %w", ctx.Status(http.StatusBadRequest).Error("bad"))
		}))

		w := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "bad\n", w.Body.String())
	})

	t.Run("helper Error stops a second write", func(t *testing.T) {
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			if err := ctx.Status(http.StatusForbidden).Error("forbidden"); err != nil {
				return err
			}
			return ctx.String("page")
		}))

		w := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		})
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "forbidden\n", w.Body.String())
	})

	t.Run("panic value is the returned error", func(t *testing.T) {
		wantErr := fmt.Errorf("boom")
		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return wantErr
		}))

		defer func() {
			assert.Equal(t, wantErr, recover())
		}()
		invokeHandler(app, "GET", "/", nil)
		t.Fatal("expected panic")
	})
}
