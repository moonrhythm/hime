package hime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	t.Parallel()

	app := New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := newContext(app, w, r)

	assert.Equal(t, r, ctx.Request())
	assert.Equal(t, w, ctx.ResponseWriter())
	ctx.Status(500)
	assert.Equal(t, 500, ctx.code)
	assert.Equal(t, &Param{Name: "a", Value: 1}, ctx.Param("a", 1))

	nctx := context.WithValue(ctx, "a", "b")
	ctx.WithContext(nctx)
	assert.Equal(t, nctx, ctx.Request().Context())
	assert.Equal(t, "b", ctx.Value("a"))

	ctx.WithValue("t", "p")
	assert.Equal(t, "p", ctx.Value("t"))

	nr := httptest.NewRequest(http.MethodPost, "/test", nil)
	ctx.WithRequest(nr)
	assert.Equal(t, nr, ctx.Request())

	nw := httptest.NewRecorder()
	ctx.WithResponseWriter(nw)
	assert.Equal(t, nw, ctx.ResponseWriter())
}
