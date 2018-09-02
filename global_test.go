package hime

import (
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

			assert.Equal(t, "value1", app.Global("key1"))
			assert.Equal(t, "value2", app.Global("key2"))
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
			})).ServeHTTP(w, r)
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
				assert.Equal(t, "value1", ctx.Global("key1"))
				assert.Equal(t, "value2", ctx.Global("key2"))
				assert.Nil(t, ctx.Global("key3"))
				return nil
			})).ServeHTTP(w, r)
		})
	})
}
