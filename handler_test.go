package hime

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	t.Run("panic on error", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(H(func(ctx *Context) error {
				return fmt.Errorf("must panic")
			}))

		assert.Panics(t, func() {
			invokeHandler(app, "GET", "/", nil)
		})
	})

	t.Run("net/http", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("ok"))
			}))

		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/", nil, "ok")
	})

	t.Run("hime", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(H(func(ctx *Context) error {
				return ctx.String("ok")
			}))

		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/", nil, "ok")
	})

	t.Run("default handler", func(t *testing.T) {
		t.Parallel()

		app := New()
		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/", nil, "404")
	})
}
