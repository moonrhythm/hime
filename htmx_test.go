package hime_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moonrhythm/hime"
)

func newHTMXContext(htmx bool) (*hime.Context, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if htmx {
		r.Header.Set("HX-Request", "true")
	}
	w := httptest.NewRecorder()
	return hime.NewAppContext(hime.New(), w, r), w
}

func TestContextIsHTMX(t *testing.T) {
	t.Parallel()

	ctx, _ := newHTMXContext(true)
	assert.True(t, ctx.IsHTMX())

	ctx, _ = newHTMXContext(false)
	assert.False(t, ctx.IsHTMX())
}

func TestContextHTMXRedirect(t *testing.T) {
	t.Parallel()

	ctx, w := newHTMXContext(true)
	assert.NoError(t, ctx.HTMXRedirect("/dashboard"))
	assert.Equal(t, "/dashboard", w.Header().Get("HX-Redirect"))

	// params are applied like Redirect
	ctx, w = newHTMXContext(true)
	assert.NoError(t, ctx.HTMXRedirect("/items", &hime.Param{Name: "page", Value: 2}))
	assert.Equal(t, "/items?page=2", w.Header().Get("HX-Redirect"))
}

func TestContextHTMXRefresh(t *testing.T) {
	t.Parallel()

	ctx, w := newHTMXContext(true)
	assert.NoError(t, ctx.HTMXRefresh())
	assert.Equal(t, "true", w.Header().Get("HX-Refresh"))
}

func TestContextHTMXReswapRetarget(t *testing.T) {
	t.Parallel()

	ctx, w := newHTMXContext(true)
	// chainable, composing with a render method
	assert.NoError(t, ctx.HTMXRetarget("#list").HTMXReswap("outerHTML").String("ok"))
	assert.Equal(t, "#list", w.Header().Get("HX-Retarget"))
	assert.Equal(t, "outerHTML", w.Header().Get("HX-Reswap"))
	assert.Equal(t, "ok", w.Body.String())
}

func TestContextHTMXTrigger(t *testing.T) {
	t.Parallel()

	t.Run("bare event name", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXTrigger("itemAdded")
		assert.Equal(t, "itemAdded", w.Header().Get("HX-Trigger"))
	})

	t.Run("event with detail as json", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXTrigger("itemAdded", map[string]any{"id": 7})
		assert.JSONEq(t, `{"itemAdded":{"id":7}}`, w.Header().Get("HX-Trigger"))
	})

	t.Run("too many detail args panics", func(t *testing.T) {
		t.Parallel()
		ctx, _ := newHTMXContext(true)
		assert.Panics(t, func() { ctx.HTMXTrigger("e", 1, 2) })
	})

	t.Run("unmarshalable detail panics", func(t *testing.T) {
		t.Parallel()
		ctx, _ := newHTMXContext(true)
		assert.Panics(t, func() { ctx.HTMXTrigger("e", make(chan int)) })
	})
}
