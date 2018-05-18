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
			assert.Len(t, app.globals, 1)
			assert.Equal(t, "test", app.Global("data1"))
			assert.Nil(t, app.Global("invalid"))

			// routes
			assert.Len(t, app.routes, 2)
			assert.Equal(t, "/", app.Route("index"))
			assert.Equal(t, "/about", app.Route("about"))

			// templates
			assert.Len(t, app.template, 2)
			assert.Contains(t, app.template, "main")
			assert.Contains(t, app.template, "main2")

			// server
			assert.Equal(t, ":8080", app.Addr)
			assert.Equal(t, 10*time.Second, app.ReadTimeout)
			assert.Equal(t, 5*time.Second, app.ReadHeaderTimeout)
			assert.Equal(t, 6*time.Second, app.WriteTimeout)
			assert.Equal(t, 30*time.Second, app.IdleTimeout)

			// graceful
			assert.NotNil(t, app.gracefulShutdown)
			assert.Equal(t, time.Minute, app.gracefulShutdown.timeout)
			assert.Equal(t, 5*time.Second, app.gracefulShutdown.wait)
		})
	})

	t.Run("Config2", func(t *testing.T) {
		assert.NotPanics(t, func() {
			app := New().ParseConfigFile("testdata/config2.yaml")

			assert.Len(t, app.globals, 0)
			assert.Len(t, app.routes, 0)
			assert.Len(t, app.template, 0)

			// server
			assert.Empty(t, app.ReadTimeout)
			assert.Empty(t, app.ReadHeaderTimeout)
			assert.Empty(t, app.WriteTimeout)
			assert.Empty(t, app.IdleTimeout)

			// graceful
			assert.NotNil(t, app.gracefulShutdown)
			assert.Empty(t, app.gracefulShutdown.timeout)
			assert.Empty(t, app.gracefulShutdown.wait)
		})
	})

	t.Run("Config3", func(t *testing.T) {
		assert.NotPanics(t, func() {
			app := New().ParseConfigFile("testdata/config3.yaml")

			assert.Len(t, app.globals, 0)
			assert.Len(t, app.routes, 0)
			assert.Len(t, app.template, 0)

			// server
			assert.Empty(t, app.ReadTimeout)
			assert.Empty(t, app.ReadHeaderTimeout)
			assert.Empty(t, app.WriteTimeout)
			assert.Empty(t, app.IdleTimeout)

			// graceful
			assert.Nil(t, app.gracefulShutdown)
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
