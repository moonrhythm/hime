package hime

import (
	"testing"

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
		})
	})

	t.Run("Config2", func(t *testing.T) {
		assert.NotPanics(t, func() {
			app := New().ParseConfigFile("testdata/config2.yaml")

			assert.Equal(t, mapLen(&app.globals), 0)
			assert.Len(t, app.routes, 0)
			assert.Len(t, app.template, 0)
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
