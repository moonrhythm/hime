package hime_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moonrhythm/hime"
)

// lastSetCookie returns the last Set-Cookie with the given name, matching how a
// browser keeps the last value when a name is set more than once.
func lastSetCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	var c *http.Cookie
	for _, x := range w.Result().Cookies() {
		if x.Name == name {
			c = x
		}
	}
	return c
}

func TestContextFlash(t *testing.T) {
	t.Parallel()

	// request 1: queue flashes, then (in real life) redirect
	w := httptest.NewRecorder()
	ctx := hime.NewAppContext(hime.New(), w, httptest.NewRequest(http.MethodPost, "/save", nil))
	ctx.AddFlash("success", "Saved")
	ctx.AddFlash("success", "Done")
	ctx.AddFlash("error", "Careful")

	c := lastSetCookie(w, "flash")
	if !assert.NotNil(t, c) {
		return
	}

	// request 2: read + clear
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.AddCookie(c)
	ctx2 := hime.NewAppContext(hime.New(), w2, r2)

	assert.Equal(t, map[string][]string{
		"success": {"Saved", "Done"},
		"error":   {"Careful"},
	}, ctx2.Flashes())

	// reading clears the cookie
	cleared := lastSetCookie(w2, "flash")
	if assert.NotNil(t, cleared) {
		assert.True(t, cleared.MaxAge < 0 || cleared.Value == "")
	}
}

func TestContextFlashesEmpty(t *testing.T) {
	t.Parallel()

	ctx := hime.NewAppContext(hime.New(), httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Nil(t, ctx.Flashes())
}

func TestContextFlashesCorrupt(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "flash", Value: "!!!not-base64!!!"})
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.Nil(t, ctx.Flashes())
	assert.NotNil(t, lastSetCookie(w, "flash")) // bad cookie is still cleared
}

func TestContextFlashesInvalidJSON(t *testing.T) {
	t.Parallel()

	// valid base64 but not valid JSON
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "flash", Value: base64.RawURLEncoding.EncodeToString([]byte("notjson"))})
	ctx := hime.NewAppContext(hime.New(), w, r)

	assert.Nil(t, ctx.Flashes())
}
