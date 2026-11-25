package hime

import (
	"bytes"
	"context"
	"crypto/tls"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	t.Parallel()

	t.Run("TemplateFuncs", func(t *testing.T) {
		app := New()
		app.TemplateFunc("a", func() string {
			return "a1"
		})

		var buf bytes.Buffer
		tmpl := template.Must(app.parent.New("").Parse("{{a}}"))
		err := tmpl.Execute(&buf, nil)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, "a1", buf.String()) {
			return
		}

		app.TemplateFuncs(template.FuncMap{
			"a": func() string { return "a2" },
			"b": func() string { return "b2" },
			"c": func() string { return "c2" },
		})
		buf.Reset()
		tmpl = template.Must(app.parent.New("").Parse("{{a}}{{b}}{{c}}"))
		err = tmpl.Execute(&buf, nil)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, "a2b2c2", buf.String()) {
			return
		}
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
		app.HTMXAwareRedirect = true

		app2 := app.Clone()
		assert.NotNil(t, app2)

		app2.Routes(Routes{
			"a": "4",
		})
		app2.Globals(Globals{
			"q": "p",
		})

		assert.NotEqual(t, app, app2)
		assert.NotEqual(t, app.routes, app2.routes)
		assert.Equal(t, app.Global("q"), "z")
		assert.Equal(t, app2.Global("q"), "p")
		assert.NotNil(t, app2.srv.TLSConfig)
		assert.True(t, app2.HTMXAwareRedirect)
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
		defer app.Shutdown()

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
		defer app.Shutdown()

		time.Sleep(time.Second)

		client := http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
		_, err := client.Get("https://localhost:8082")
		assert.NoError(t, err)
		assert.True(t, called)
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

func TestAppCloneServerSettings(t *testing.T) {
	t.Parallel()

	app := New()
	logger := log.New(io.Discard, "", 0)
	trust := func(*http.Request) bool { return true }
	connState := func(net.Conn, http.ConnState) {}
	baseCtx := func(net.Listener) context.Context { return context.Background() }

	app.srv.Addr = ":9090"
	app.srv.ReadTimeout = time.Second
	app.srv.ReadHeaderTimeout = 2 * time.Second
	app.srv.WriteTimeout = 3 * time.Second
	app.srv.IdleTimeout = 4 * time.Second
	app.srv.MaxHeaderBytes = 1234
	app.srv.TCPKeepAlivePeriod = 5 * time.Second
	app.srv.GraceTimeout = 6 * time.Second
	app.srv.WaitBeforeShutdown = 7 * time.Second
	app.srv.ErrorLog = logger
	app.srv.TrustProxy = trust
	app.srv.H2C = true
	app.srv.ReusePort = true
	app.srv.ConnState = connState
	app.srv.TLSConfig = Restricted()
	app.srv.BaseContext = baseCtx
	app.Handler(Handler(func(ctx *Context) error {
		return ctx.String("cloned")
	}))

	cloned := app.Clone()
	s := cloned.Server()
	assert.Equal(t, ":9090", s.Addr)
	assert.Equal(t, time.Second, s.ReadTimeout)
	assert.Equal(t, 2*time.Second, s.ReadHeaderTimeout)
	assert.Equal(t, 3*time.Second, s.WriteTimeout)
	assert.Equal(t, 4*time.Second, s.IdleTimeout)
	assert.Equal(t, 1234, s.MaxHeaderBytes)
	assert.Equal(t, 5*time.Second, s.TCPKeepAlivePeriod)
	assert.Equal(t, 6*time.Second, s.GraceTimeout)
	assert.Equal(t, 7*time.Second, s.WaitBeforeShutdown)
	assert.Same(t, logger, s.ErrorLog)
	assert.True(t, s.TrustProxy(nil))
	assert.True(t, s.H2C)
	assert.True(t, s.ReusePort)
	assert.NotNil(t, s.ConnState)
	assert.Equal(t, Restricted().CurvePreferences, s.TLSConfig.CurvePreferences)
	assert.Equal(t, Restricted().CipherSuites, s.TLSConfig.CipherSuites)
	assert.NotNil(t, s.BaseContext)
	assert.Same(t, cloned, s.Handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	cloned.ServeHTTP(w, r)
	assert.Equal(t, "cloned", w.Body.String())

	// Source has not served, so this handler is the one its first request runs.
	app.Handler(Handler(func(ctx *Context) error {
		return ctx.String("base")
	}))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, "base", w.Body.String())
}

func TestAppSetServerNil(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() { New().SetServer(nil) })
}

func TestAppCloneNilTLSConfig(t *testing.T) {
	t.Parallel()

	// A fresh app has a nil TLSConfig; cloning must not panic and must
	// preserve the nil TLSConfig (crypto/tls Clone handles nil receivers).
	app := New()
	assert.Nil(t, app.Server().TLSConfig)

	var cloned *App
	assert.NotPanics(t, func() { cloned = app.Clone() })
	assert.Nil(t, cloned.Server().TLSConfig)
}
