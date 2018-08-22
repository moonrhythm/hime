package hime

import (
	"html/template"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	t.Parallel()

	t.Run("TemplateFuncs", func(t *testing.T) {
		app := New()
		app.TemplateFunc("a", func() {})
		assert.Len(t, app.templateFuncs, 1)

		app.TemplateFuncs(template.FuncMap{
			"a": func() {},
			"b": func() {},
			"c": func() {},
		})
		assert.Len(t, app.templateFuncs, 2)
	})

	t.Run("GracefulShutdown", func(t *testing.T) {
		app := New()
		assert.Nil(t, app.gs)
		gs := app.GracefulShutdown()
		assert.NotNil(t, app.gs)
		assert.Equal(t, app.gs, gs)

		gs.Timeout(10 * time.Second)
		assert.Equal(t, 10*time.Second, gs.timeout)

		gs.Wait(5 * time.Second)
		assert.Equal(t, 5*time.Second, gs.wait)

		gs.Notify(func() {})
		gs.Notify(func() {})
		gs.Notify(func() {})
		assert.Len(t, gs.notiFns, 3)
	})

	t.Run("Address", func(t *testing.T) {
		app := New()
		app.Address(":1234")
		assert.Equal(t, ":1234", app.srv.Addr)
	})

	t.Run("Clone", func(t *testing.T) {
		app := New()
		app.Routes(Routes{
			"a": "1",
			"b": "2",
		})
		app.Globals(Globals{
			"q": "z",
			"w": "x",
		})
		app.srv.TLSConfig = Compatible()
		app.GracefulShutdown()

		app2 := app.Clone()
		assert.NotNil(t, app2)

		app2.Routes(Routes{
			"a": "4",
		}).Globals(Globals{
			"q": "p",
		})
		app2.GracefulShutdown().Wait(10 * time.Second)

		assert.NotEqual(t, app, app2)
		assert.NotEqual(t, app.routes, app2.routes)
		assert.NotEqual(t, app.globals, app2.globals)
		assert.NotEqual(t, app.gs, app2.gs)
		assert.NotNil(t, app2.srv.TLSConfig)
	})

	t.Run("SelfSign empty param", func(t *testing.T) {
		app := New()
		app.SelfSign(SelfSign{})

		assert.NotNil(t, app.srv.TLSConfig)
	})

	t.Run("SelfSign invalid param", func(t *testing.T) {
		app := New()
		opt := SelfSign{}
		opt.Key.Algo = "invalid"

		assert.Panics(t, func() { app.SelfSign(opt) })
	})

	t.Run("Server", func(t *testing.T) {
		app := New()
		assert.NotNil(t, app.Server())
	})
}
