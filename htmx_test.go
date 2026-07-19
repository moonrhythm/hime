package hime_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func newHTMXContextHeaders(headers map[string]string) (*hime.Context, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	return hime.NewAppContext(hime.New(), w, r), w
}

func varyTokens(w *httptest.ResponseRecorder) map[string]bool {
	m := make(map[string]bool)
	for _, v := range w.Header().Values("Vary") {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				m[strings.ToLower(part)] = true
			}
		}
	}
	return m
}

func assertHTMXVary(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	tokens := varyTokens(w)
	assert.True(t, tokens["hx-request"], "Vary missing HX-Request: %v", w.Header().Values("Vary"))
	assert.True(t, tokens["hx-boosted"], "Vary missing HX-Boosted: %v", w.Header().Values("Vary"))
	assert.True(t, tokens["hx-history-restore-request"], "Vary missing HX-History-Restore-Request: %v", w.Header().Values("Vary"))
}

func TestContextIsHTMX(t *testing.T) {
	t.Parallel()

	ctx, w := newHTMXContext(true)
	assert.True(t, ctx.IsHTMX())
	assert.Empty(t, w.Header().Get("Vary"), "predicates must not mutate the response")

	ctx, w = newHTMXContext(false)
	assert.False(t, ctx.IsHTMX())
	assert.Empty(t, w.Header().Get("Vary"))
}

func TestContextIsBoosted(t *testing.T) {
	t.Parallel()

	ctx, w := newHTMXContextHeaders(map[string]string{
		"HX-Request": "true",
		"HX-Boosted": "true",
	})
	assert.True(t, ctx.IsBoosted())
	assert.Empty(t, w.Header().Get("Vary"), "predicates must not mutate the response")

	ctx, w = newHTMXContext(true)
	assert.False(t, ctx.IsBoosted())
	assert.Empty(t, w.Header().Get("Vary"))
}

func TestContextWantsPartial(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		headers map[string]string
		want    bool
	}{
		{"plain", nil, false},
		{"htmx", map[string]string{"HX-Request": "true"}, true},
		{"boosted", map[string]string{"HX-Request": "true", "HX-Boosted": "true"}, false},
		{"history-restore", map[string]string{
			"HX-Request":                 "true",
			"HX-History-Restore-Request": "true",
		}, false},
		{"boosted history-restore", map[string]string{
			"HX-Request":                 "true",
			"HX-Boosted":                 "true",
			"HX-History-Restore-Request": "true",
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, w := newHTMXContextHeaders(tc.headers)
			assert.Equal(t, tc.want, ctx.WantsPartial())
			assert.Empty(t, w.Header().Get("Vary"), "predicates must not mutate the response")
		})
	}
}

func TestContextVaryHTMX(t *testing.T) {
	t.Parallel()

	t.Run("emitted once with all three values", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		assert.Same(t, ctx, ctx.VaryHTMX())
		ctx.VaryHTMX()
		ctx.VaryHTMX()
		assertHTMXVary(t, w)

		// exactly one Vary entry containing the three tokens (no duplicates)
		var count int
		for _, v := range w.Header().Values("Vary") {
			for _, part := range strings.Split(v, ",") {
				if strings.EqualFold(strings.TrimSpace(part), "HX-Request") {
					count++
				}
			}
		}
		assert.Equal(t, 1, count)
	})

	t.Run("appended to pre-existing Vary", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.AddHeader("Vary", "Accept-Encoding")
		ctx.VaryHTMX()
		tokens := varyTokens(w)
		assert.True(t, tokens["accept-encoding"])
		assertHTMXVary(t, w)
	})

	t.Run("present on ETag 304", func(t *testing.T) {
		t.Parallel()
		app := hime.New()
		app.ETag = true
		tmpl := app.Template()
		tmpl.Dir("testdata")
		tmpl.Root("root")
		tmpl.ParseFiles("index", "hello.tmpl")

		// first request captures ETag
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("HX-Request", "true")
		ctx := hime.NewAppContext(app, w, r)
		assert.NoError(t, ctx.VaryHTMX().View("index", nil))
		etag := w.Header().Get("ETag")
		assert.NotEmpty(t, etag)
		assertHTMXVary(t, w)

		// 304 revalidation still carries Vary
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("HX-Request", "true")
		r.Header.Set("If-None-Match", etag)
		ctx = hime.NewAppContext(app, w, r)
		assert.NoError(t, ctx.VaryHTMX().View("index", nil))
		assert.Equal(t, http.StatusNotModified, w.Code)
		assertHTMXVary(t, w)
	})
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

func TestContextHTMXPushReplaceURL(t *testing.T) {
	t.Parallel()

	t.Run("PushURL with params", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXPushURL("/items", &hime.Param{Name: "page", Value: 2})
		assert.Equal(t, "/items?page=2", w.Header().Get("HX-Push-Url"))
	})

	t.Run("ReplaceURL", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXReplaceURL("/settings")
		assert.Equal(t, "/settings", w.Header().Get("HX-Replace-Url"))
	})

	t.Run("NoPush", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXNoPush()
		assert.Equal(t, "false", w.Header().Get("HX-Push-Url"))
	})

	t.Run("Reselect", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXReselect("#main")
		assert.Equal(t, "#main", w.Header().Get("HX-Reselect"))
	})
}

func TestContextHTMXLocation(t *testing.T) {
	t.Parallel()

	ctx, w := newHTMXContext(true)
	assert.NoError(t, ctx.HTMXLocation("/dashboard", &hime.Param{Name: "tab", Value: "a"}))
	assert.Equal(t, "/dashboard?tab=a", w.Header().Get("HX-Location"))
	assert.Equal(t, http.StatusNoContent, w.Code)
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

	t.Run("merge bare and detail", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXTrigger("showToast")
		ctx.HTMXTrigger("itemAdded", map[string]any{"id": 7})
		assert.JSONEq(t, `{"showToast":null,"itemAdded":{"id":7}}`, w.Header().Get("HX-Trigger"))
	})

	t.Run("merge multiple bare", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXTrigger("a")
		ctx.HTMXTrigger("b")
		assert.Equal(t, "a, b", w.Header().Get("HX-Trigger"))
	})

	t.Run("three timings independent", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXTrigger("t1")
		ctx.HTMXTriggerAfterSwap("t2", "d2")
		ctx.HTMXTriggerAfterSettle("t3")
		assert.Equal(t, "t1", w.Header().Get("HX-Trigger"))
		assert.JSONEq(t, `{"t2":"d2"}`, w.Header().Get("HX-Trigger-After-Swap"))
		assert.Equal(t, "t3", w.Header().Get("HX-Trigger-After-Settle"))
	})

	t.Run("survives Redirect", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXTrigger("afterSave")
		assert.NoError(t, ctx.Redirect("/done"))
		assert.Equal(t, "afterSave", w.Header().Get("HX-Trigger"))
		assert.Equal(t, http.StatusFound, w.Code)
	})

	t.Run("survives NoContent", func(t *testing.T) {
		t.Parallel()
		ctx, w := newHTMXContext(true)
		ctx.HTMXTrigger("done")
		assert.NoError(t, ctx.NoContent())
		assert.Equal(t, "done", w.Header().Get("HX-Trigger"))
		assert.Equal(t, http.StatusNoContent, w.Code)
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
