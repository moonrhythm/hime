package hime_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acoshift/hime"
	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	t.Run("NotPassFromApp", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Panics(t, func() { hime.NewContext(w, r) })
	})
}
