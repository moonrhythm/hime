package hime

import (
	"bytes"
	"crypto/tls"
	"html/template"
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
