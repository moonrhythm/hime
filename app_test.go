package hime

import (
	"crypto/tls"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	t.Parallel()

	app := New()

	t.Run("TemplateFuncs", func(t *testing.T) {
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
		assert.Nil(t, app.gracefulShutdown)
		gs := app.GracefulShutdown()
		assert.NotNil(t, app.gracefulShutdown)
		assert.Equal(t, app.gracefulShutdown, gs.gracefulShutdown)

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
		app.Address(":1234")
		assert.Equal(t, ":1234", app.Addr)
	})
}

func TestConfigServer(t *testing.T) {
	t.Parallel()

	app := &App{
		Addr:              ":8080",
		TLSConfig:         &tls.Config{},
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 6 * time.Second,
		WriteTimeout:      7 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1024,
		TLSNextProto:      map[string]func(*http.Server, *tls.Conn, http.Handler){},
		ConnState:         func(net.Conn, http.ConnState) {},
		ErrorLog:          log.New(os.Stderr, "", log.LstdFlags),
	}

	assert.Empty(t, &app.srv)
	app.configServer()
	assert.NotEmpty(t, &app.srv)
	assert.Equal(t, ":8080", app.srv.Addr)
}

func TestAppClone(t *testing.T) {
	t.Parallel()

	app := New().
		Routes(Routes{
			"a": "1",
			"b": "2",
		}).
		Globals(Globals{
			"q": "z",
			"w": "x",
		})

	app2 := app.Clone()
	assert.NotNil(t, app2)

	app2.Routes(Routes{
		"a": "4",
	}).Globals(Globals{
		"q": "p",
	})

	assert.NotEqual(t, app, app2)
	assert.NotEqual(t, app.routes, app2.routes)
	assert.NotEqual(t, app.globals, app2.globals)
}
