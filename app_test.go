package hime

import (
	"context"
	"crypto/tls"
	"html/template"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/http2"
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
		assert.Equal(t, gs, app.gs)

		gs.Timeout(10 * time.Second)
		assert.Equal(t, gs.timeout, 10*time.Second)

		gs.Wait(5 * time.Second)
		assert.Equal(t, gs.wait, 5*time.Second)

		gs.Notify(func() {})
		gs.Notify(func() {})
		gs.Notify(func() {})
		assert.Len(t, gs.notiFns, 3)
	})

	t.Run("Address", func(t *testing.T) {
		app := New()
		app.Address(":1234")
		assert.Equal(t, app.srv.Addr, ":1234")
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
		assert.Equal(t, app.Global("q"), "z")
		assert.Equal(t, app2.Global("q"), "p")
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

	t.Run("TLS", func(t *testing.T) {
		app := New()

		if assert.NotPanics(t, func() { app.TLS("testdata/server.crt", "testdata/server.key") }) {
			assert.NotNil(t, app.srv.TLSConfig)
		}
	})

	t.Run("TLS invalid", func(t *testing.T) {
		app := New()

		assert.Panics(t, func() { app.TLS("testdata/server.key", "testdata/server.crt") })
	})

	t.Run("ListenAndServe", func(t *testing.T) {
		t.Parallel()

		app := New()
		called := false
		app.Handler(Handler(func(ctx *Context) error {
			called = true
			return ctx.String("Hello")
		}))
		app.Address(":8081")

		go app.ListenAndServe()
		defer app.Shutdown(context.Background())

		time.Sleep(time.Second)

		_, err := http.Get("http://localhost:8081")
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("ListenAndServe with tls", func(t *testing.T) {
		t.Parallel()

		app := New()
		app.TLS("testdata/server.crt", "testdata/server.key")

		called := false
		app.Handler(Handler(func(ctx *Context) error {
			called = true
			return ctx.String("Hello")
		}))
		app.Address(":8082")

		go app.ListenAndServe()
		defer app.Shutdown(context.Background())

		time.Sleep(time.Second)

		client := http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
		_, err := client.Get("https://localhost:8082")
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("ListenAndServe with graceful shutdown", func(t *testing.T) {
		t.Parallel()

		app := New()
		app.Handler(Handler(func(ctx *Context) error {
			return ctx.String("Hello")
		}))
		app.Address(":8083")

		gs := app.GracefulShutdown()
		gs.Wait(time.Second)
		gs.Timeout(time.Second)
		called := false
		gs.Notify(func() {
			called = true
		})

		go app.ListenAndServe()
		time.Sleep(time.Second)

		http.Get("http://localhost:8083")
		app.Shutdown(context.Background())
		time.Sleep(time.Second)
		assert.True(t, called)
	})

	t.Run("ListenAndServe with graceful shutdown and tls", func(t *testing.T) {
		t.Parallel()

		app := New()
		app.TLS("testdata/server.crt", "testdata/server.key")
		app.Handler(Handler(func(ctx *Context) error {
			return ctx.String("Hello")
		}))
		app.Address(":8084")

		gs := app.GracefulShutdown()
		gs.Wait(time.Second)
		gs.Timeout(time.Second)
		called := false
		gs.Notify(func() {
			called = true
		})

		go app.ListenAndServe()
		time.Sleep(time.Second)

		(&http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}).Get("https://localhost:8084")
		app.Shutdown(context.Background())
		time.Sleep(time.Second)
		assert.True(t, called)
	})

	t.Run("H2C", func(t *testing.T) {
		t.Parallel()

		app := New()
		app.H2C = true
		called := false
		app.Handler(Handler(func(ctx *Context) error {
			called = true
			assert.Equal(t, "HTTP/2.0", ctx.Request.Proto)
			return ctx.String("Hello")
		}))
		app.Address(":8085")

		go app.ListenAndServe()
		defer app.Shutdown(context.Background())

		time.Sleep(time.Second)

		client := http.Client{
			Transport: &http2.Transport{
				AllowHTTP:          true,
				DisableCompression: true,
				DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		}
		_, err := client.Get("http://localhost:8085")
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Disable H2C", func(t *testing.T) {
		t.Parallel()

		app := New()
		app.H2C = false
		app.Handler(Handler(func(ctx *Context) error {
			return ctx.String("Hello")
		}))
		app.Address(":8086")

		go app.ListenAndServe()
		defer app.Shutdown(context.Background())

		time.Sleep(time.Second)

		client := http.Client{
			Transport: &http2.Transport{
				AllowHTTP:          true,
				DisableCompression: true,
				DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		}
		_, err := client.Get("http://localhost:8086")
		assert.Error(t, err)

		_, err = http.Get("http://localhost:8086")
		assert.NoError(t, err)
	})

	t.Run("ServeHandler", func(t *testing.T) {
		t.Parallel()

		app := New()
		called := false
		h := app.ServeHandler(Handler(func(ctx *Context) error {
			called = true
			return ctx.String("Hello")
		}))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(w, r)

		assert.True(t, called)
		assert.Equal(t, "Hello", w.Body.String())
	})
}
