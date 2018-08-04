package hime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveComma(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input  string
		Output string
	}{
		{"", ""},
		{"123", "123"},
		{"12,345", "12345"},
		{" 12, ,, 34,5, ,,", " 12  345 "},
		{"12,345.67", "12345.67"},
	}

	for _, c := range cases {
		assert.Equal(t, c.Output, removeComma(c.Input))
	}
}

func TestRequest(t *testing.T) {
	t.Run("Method", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				assert.Equal(t, ctx.r.Method, ctx.Method())
				return nil
			}))

		invokeHandler(app, "GET", "/", nil)
		invokeHandler(app, "POST", "/", nil)
		invokeHandler(app, "PUT", "/", nil)
		invokeHandler(app, "DELETE", "/", nil)
		invokeHandler(app, "PURGE", "/", nil)
		invokeHandler(app, "AAA", "/", nil)
	})

	t.Run("Query", func(t *testing.T) {
		t.Parallel()

		app := New().
			Handler(Handler(func(ctx *Context) error {
				assert.Equal(t, "a", ctx.Query().Get("t"))
				return nil
			}))

		invokeHandler(app, "GET", "/?t=a", nil)
	})
}
