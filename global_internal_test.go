package hime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobal(t *testing.T) {
	t.Run("clone nil", func(t *testing.T) {
		assert.Nil(t, cloneGlobals(nil))
	})

	t.Run("clone empty", func(t *testing.T) {
		r := Globals{}
		assert.Equal(t, Globals{}, cloneGlobals(r))
	})

	t.Run("clone data", func(t *testing.T) {
		r := Globals{
			"a": 1,
			"b": 2,
		}

		p := cloneGlobals(r)
		assert.Equal(t, r, p)

		p["a"] = 5
		assert.NotEqual(t, r, p)
	})
}
