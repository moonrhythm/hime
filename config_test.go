package hime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("Config1", func(t *testing.T) {
		assert.NotPanics(t, func() {
			app := New()
			app.ParseConfigFile("testdata/config1.yaml")

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
			app := New()
			app.ParseConfigFile("testdata/config2.yaml")

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

func TestConfigMerge(t *testing.T) {
	t.Parallel()

	// Multiple Config calls accumulate; colliding keys take the latest value.
	app := New()
	app.Config(AppConfig{
		Globals: Globals{"a": "1"},
		Routes:  Routes{"r1": "/1"},
	})
	app.Config(AppConfig{
		Globals: Globals{"b": "2", "a": "1override"},
		Routes:  Routes{"r2": "/2"},
	})

	assert.Equal(t, "1override", app.Global("a"))
	assert.Equal(t, "2", app.Global("b"))
	assert.Equal(t, "/1", app.Route("r1"))
	assert.Equal(t, "/2", app.Route("r2"))
}

func TestConfigParseNilData(t *testing.T) {
	t.Parallel()

	app := New()
	assert.NotPanics(t, func() { app.ParseConfig(nil) })
	assert.Equal(t, 0, mapLen(&app.globals))
	assert.Len(t, app.routes, 0)
	assert.Len(t, app.template, 0)
}

func TestConfigParseConfigFileEmptyName(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() { New().ParseConfigFile("") })
}
