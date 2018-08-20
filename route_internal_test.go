package hime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute(t *testing.T) {
	t.Run("clone nil", func(t *testing.T) {
		assert.Nil(t, cloneRoutes(nil))
	})

	t.Run("clone empty", func(t *testing.T) {
		r := Routes{}
		assert.Equal(t, Routes{}, cloneRoutes(r))
	})

	t.Run("clone data", func(t *testing.T) {
		r := Routes{
			"a": "/b",
			"b": "/cd",
		}

		p := cloneRoutes(r)
		assert.Equal(t, r, p)

		p["a"] = "/a"
		assert.NotEqual(t, r, p)
	})
}
