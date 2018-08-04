package hime

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobal(t *testing.T) {
	t.Parallel()

	app := New()

	assert.Nil(t, app.Global("key1"))
	assert.Nil(t, app.Global("key2"))

	app.
		Globals(Globals{
			"key1": "value1",
			"key2": "value2",
		}).
		Handler(Handler(func(ctx *Context) error {
			assert.Equal(t, "value1", ctx.Global("key1"))
			assert.Equal(t, "value2", ctx.Global("key2"))
			assert.Nil(t, ctx.Global("invalid"))
			return nil
		}))

	assert.Equal(t, "value1", app.Global("key1"))
	assert.Equal(t, "value2", app.Global("key2"))
	assert.Nil(t, app.Global("invalid"))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
}

func TestCloneGlobals(t *testing.T) {
	t.Parallel()

	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, cloneGlobals(nil))
	})

	t.Run("Normal", func(t *testing.T) {
		g := Globals{
			"a": 1,
			"b": 2,
		}

		p := cloneGlobals(g)
		p["a"] = 2
		p["c"] = 3

		assert.NotEqual(t, g, p)
	})
}
