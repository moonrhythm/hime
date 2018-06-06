package hime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute(t *testing.T) {
	t.Parallel()

	app := New()
	assert.Panics(t, func() { app.Route("empty") })

	app.Routes(Routes{"a": "/b", "b": "/cd"})
	assert.Len(t, app.routes, 2)
	assert.Equal(t, "/b", app.Route("a"))
	assert.Equal(t, "/cd", app.Route("b"))

	assert.Panics(t, func() { app.Route("notexists") })

	ctx := Context{app: app}
	assert.Equal(t, "/b", ctx.Route("a"))
}
