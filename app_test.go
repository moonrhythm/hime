package hime_test

import (
	"net/http"
	"testing"

	"github.com/acoshift/hime"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	t.Run("net/http", func(t *testing.T) {
		app := hime.New().
			Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("ok"))
			}))

		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/", nil, "ok")
	})

	t.Run("hime", func(t *testing.T) {
		app := hime.New().
			Handler(hime.H(func(ctx *hime.Context) hime.Result {
				return ctx.String("ok")
			}))

		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/", nil, "ok")
	})

	t.Run("default handler", func(t *testing.T) {
		app := hime.New()
		assert.HTTPBodyContains(t, app.ServeHTTP, "GET", "/", nil, "404")
	})
}
