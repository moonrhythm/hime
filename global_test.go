package hime

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobal(t *testing.T) {
	t.Run("App", func(t *testing.T) {
		t.Run("retrieve any data from empty global", func(t *testing.T) {
			app := New()

			assert.Nil(t, app.Global("key1"))
		})

		t.Run("retrieve data from global", func(t *testing.T) {
			app := New()
			app.Globals(Globals{
				"key1": "value1",
				"key2": "value2",
			})

			assert.Equal(t, app.Global("key1"), "value1")
			assert.Equal(t, app.Global("key2"), "value2")
			assert.Nil(t, app.Global("key3"))
		})
	})

	t.Run("Context", func(t *testing.T) {
		t.Run("retrieve any data from empty global", func(t *testing.T) {
			app := New()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			app.Handler(Handler(func(ctx *Context) error {
				assert.Nil(t, app.Global("key1"))
				return nil
			}))
			app.ServeHTTP(w, r)
		})

		t.Run("retrieve data from global", func(t *testing.T) {
			app := New()
			app.Globals(Globals{
				"key1": "value1",
				"key2": "value2",
			})

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			app.Handler(Handler(func(ctx *Context) error {
				assert.Equal(t, ctx.Global("key1"), "value1")
				assert.Equal(t, ctx.Global("key2"), "value2")
				assert.Equal(t, Global(ctx, "key1"), "value1")
				assert.Equal(t, Global(ctx, "key2"), "value2")
				assert.Nil(t, ctx.Global("key3"))
				return nil
			}))
			app.ServeHTTP(w, r)
		})
	})
}

func TestGlobalFuncWithoutApp(t *testing.T) {
	assert.PanicsWithValue(t, ErrAppNotFound, func() {
		Global(context.Background(), "key1")
	})
}

func TestGlobalsMerge(t *testing.T) {
	app := New()
	app.Globals(Globals{"k1": "v1"})
	app.Globals(Globals{"k2": "v2", "k1": "v1b"})

	assert.Equal(t, "v1b", app.Global("k1"))
	assert.Equal(t, "v2", app.Global("k2"))
}

func TestGlobalZeroValueVsMissing(t *testing.T) {
	// Global discards the "found" boolean from sync.Map.Load, so a stored
	// zero value and a missing key both surface as their natural value/nil.
	app := New()
	app.Globals(Globals{"empty": "", "zero": 0})

	assert.Equal(t, "", app.Global("empty"))
	assert.Equal(t, 0, app.Global("zero"))
	assert.Nil(t, app.Global("missing"))
}
