package hime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("Config1", func(t *testing.T) {
		assert.NotPanics(t, func() {
			app := New().ParseConfigFile("testdata/config1.yaml")

			// globals
			assert.Equal(t, mapLen(&app.globals), 1)
			assert.Equal(t, app.Global("data1"), "test")
			assert.Nil(t, app.Global("invalid"))

			// routes
			assert.Len(t, app.routes, 2)
			assert.Equal(t, app.Route("index"), "/")
			assert.Equal(t, app.Route("about"), "/about")

			// templates
			assert.Len(t, app.template, 2)
			assert.Contains(t, app.template, "main")
			assert.Contains(t, app.template, "main2")

			// server
			assert.Equal(t, app.srv.Addr, ":8080")
			assert.Equal(t, app.srv.ReadTimeout, 10*time.Second)
			assert.Equal(t, app.srv.ReadHeaderTimeout, 5*time.Second)
			assert.Equal(t, app.srv.WriteTimeout, 6*time.Second)
			assert.Equal(t, app.srv.IdleTimeout, 30*time.Second)
			assert.True(t, app.reusePort)
			assert.Len(t, app.srv.TLSConfig.Certificates, 1)

			// graceful
			assert.NotNil(t, app.gs)
			assert.Equal(t, app.gs.timeout, time.Minute)
			assert.Equal(t, app.gs.wait, 5*time.Second)
		})
	})

	t.Run("Config2", func(t *testing.T) {
		assert.NotPanics(t, func() {
			app := New().ParseConfigFile("testdata/config2.yaml")

			assert.Equal(t, mapLen(&app.globals), 0)
			assert.Len(t, app.routes, 0)
			assert.Len(t, app.template, 0)

			// server
			assert.Empty(t, app.srv.ReadTimeout)
			assert.Empty(t, app.srv.ReadHeaderTimeout)
			assert.Empty(t, app.srv.WriteTimeout)
			assert.Empty(t, app.srv.IdleTimeout)

			// graceful
			assert.NotNil(t, app.gs)
			assert.Empty(t, app.gs.timeout)
			assert.Empty(t, app.gs.wait)
		})
	})

	t.Run("Config3", func(t *testing.T) {
		assert.NotPanics(t, func() {
			app := New().ParseConfigFile("testdata/config3.yaml")

			assert.Equal(t, mapLen(&app.globals), 0)
			assert.Len(t, app.routes, 0)
			assert.Len(t, app.template, 0)

			// server
			assert.Empty(t, app.srv.ReadTimeout)
			assert.Empty(t, app.srv.ReadHeaderTimeout)
			assert.Empty(t, app.srv.WriteTimeout)
			assert.Empty(t, app.srv.IdleTimeout)

			// graceful
			assert.Nil(t, app.gs)
		})
	})

	t.Run("ConfigNotFound", func(t *testing.T) {
		assert.Panics(t, func() {
			New().ParseConfigFile("testdata/notexists.yaml")
		})
	})

	t.Run("Invalid1", func(t *testing.T) {
		assert.Panics(t, func() {
			New().ParseConfigFile("testdata/invalid1.yaml")
		})
	})

	t.Run("Invalid2", func(t *testing.T) {
		assert.Panics(t, func() {
			New().ParseConfigFile("testdata/invalid2.yaml")
		})
	})
}
