package hime_test

import (
	"testing"

	"github.com/acoshift/hime"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		assert.NotPanics(t, func() {
			app := hime.New().LoadFromFile("testdata/config1.yaml")
			assert.Equal(t, "test", app.Global("data1"))
			assert.Nil(t, app.Global("invalid"))
			assert.Equal(t, "/", app.Route("index"))
			assert.Equal(t, "/about", app.Route("about"))
		})
	})

	t.Run("ConfigNotFound", func(t *testing.T) {
		assert.Panics(t, func() {
			hime.New().LoadFromFile("testdata/notexists.yaml")
		})
	})

	t.Run("Invalid1", func(t *testing.T) {
		assert.Panics(t, func() {
			hime.New().LoadFromFile("testdata/invalid1.yaml")
		})
	})

	t.Run("Invalid2", func(t *testing.T) {
		assert.Panics(t, func() {
			hime.New().LoadFromFile("testdata/invalid2.yaml")
		})
	})
}
