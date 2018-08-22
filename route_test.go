package hime

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute(t *testing.T) {
	t.Run("clone nil", func(t *testing.T) {
		assert.Nil(t, cloneRoutes(nil))
	})

	t.Run("clone empty", func(t *testing.T) {
		r := Routes{}
		assert.Equal(t, Routes{}, cloneRoutes(r))
	})

	t.Run("clone data", func(t *testing.T) {
		r := Routes{
			"a": "/b",
			"b": "/cd",
		}

		p := cloneRoutes(r)
		assert.Equal(t, r, p)

		p["a"] = "/a"
		assert.NotEqual(t, r, p)
	})

	t.Run("App", func(t *testing.T) {
		t.Run("panic when retrieve route from empty route app", func(t *testing.T) {
			app := New()

			assert.Panics(t, func() { app.Route("r1") })
		})

		t.Run("retrieve valid route", func(t *testing.T) {
			app := New()
			app.Routes(Routes{
				"a": "/b",
				"b": "/cd",
			})

			assert.Equal(t, "/b", app.Route("a"))
			assert.Equal(t, "/cd", app.Route("b"))
		})
	})

	t.Run("panic when retrieve not exists route", func(t *testing.T) {
		app := New()
		app.Routes(Routes{
			"a": "/b",
			"b": "/cd",
		})

		assert.Panics(t, func() { app.Route("c") })
	})

	t.Run("Context", func(t *testing.T) {
		t.Run("retrieve route from context", func(t *testing.T) {
			app := New()
			app.Routes(Routes{
				"a": "/b",
				"b": "/cd",
			})

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			app.Handler(Handler(func(ctx *Context) error {
				assert.Equal(t, "/b", ctx.Route("a"))
				assert.Equal(t, "/cd", ctx.Route("b"))
				return nil
			})).ServeHTTP(w, r)
		})

		t.Run("panic when retrieve not exists route", func(t *testing.T) {
			app := New()
			app.Routes(Routes{
				"a": "/b",
				"b": "/cd",
			})

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			app.Handler(Handler(func(ctx *Context) error {
				assert.Panics(t, func() { ctx.Route("c") })
				return nil
			})).ServeHTTP(w, r)
		})
	})
}
