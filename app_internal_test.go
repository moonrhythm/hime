package hime

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	t.Parallel()

	app := New()

	t.Run("BeforeRender", func(t *testing.T) {
		m := func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		}

		app.BeforeRender(m)
		assert.Equal(t, reflect.ValueOf(m).Pointer(), reflect.ValueOf(app.beforeRender).Pointer())
	})

	t.Run("Routes", func(t *testing.T) {
		app.Routes(Routes{"a": "/b", "b": "/cd"})
		assert.Len(t, app.routes, 2)
		assert.Equal(t, "/b", app.Route("a"))
		assert.Equal(t, "/cd", app.Route("b"))
	})

	t.Run("Globals", func(t *testing.T) {
		app.Globals(Globals{"a": 12, "b": "34"})
		assert.Len(t, app.globals, 2)
		assert.Equal(t, 12, app.Global("a"))
		assert.Equal(t, "34", app.Global("b"))
	})

	t.Run("GracefulShutdown", func(t *testing.T) {
		assert.Nil(t, app.gracefulShutdown)
		gs := app.GracefulShutdown()
		assert.NotNil(t, app.gracefulShutdown)
		assert.Equal(t, app.gracefulShutdown, gs.gracefulShutdown)

		gs.Timeout(10 * time.Second)
		assert.Equal(t, 10*time.Second, gs.timeout)

		gs.Wait(5 * time.Second)
		assert.Equal(t, 5*time.Second, gs.wait)

		gs.Before(func() {})
		gs.Before(func() {})
		assert.Len(t, gs.beforeFns, 2)

		gs.Notify(func() {})
		gs.Notify(func() {})
		gs.Notify(func() {})
		assert.Len(t, gs.notiFns, 3)
	})
}
