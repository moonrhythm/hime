package hime_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moonrhythm/hime"
)

var testCookieKey = []byte("0123456789abcdef0123456789abcdef") // 32 bytes

// flipFirstByte returns s with its first byte changed, to simulate tampering.
func flipFirstByte(s string) string {
	b := []byte(s)
	if b[0] == 'A' {
		b[0] = 'B'
	} else {
		b[0] = 'A'
	}
	return string(b)
}

// readSetCookie returns the cookie with the given name from a recorded response.
func readSetCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestHMACCookieSigner(t *testing.T) {
	t.Parallel()

	signer := hime.NewHMACCookieSigner(testCookieKey)

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()
		signed, err := signer.Sign("session", "user42")
		assert.NoError(t, err)
		assert.NotEqual(t, "user42", signed) // value is encoded + signed

		got, err := signer.Verify("session", signed)
		assert.NoError(t, err)
		assert.Equal(t, "user42", got)
	})

	t.Run("empty value round trips", func(t *testing.T) {
		t.Parallel()
		signed, err := signer.Sign("session", "")
		assert.NoError(t, err)

		got, err := signer.Verify("session", signed)
		assert.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("binary-ish value round trips", func(t *testing.T) {
		t.Parallel()
		value := "a=b&c=d|e.f/g+h\x00\xff"
		signed, err := signer.Sign("data", value)
		assert.NoError(t, err)

		got, err := signer.Verify("data", signed)
		assert.NoError(t, err)
		assert.Equal(t, value, got)
	})

	t.Run("tampered value fails", func(t *testing.T) {
		t.Parallel()
		signed, _ := signer.Sign("session", "user42")
		_, err := signer.Verify("session", flipFirstByte(signed))
		assert.ErrorIs(t, err, hime.ErrInvalidCookie)
	})

	t.Run("wrong key fails", func(t *testing.T) {
		t.Parallel()
		signed, _ := signer.Sign("session", "user42")
		other := hime.NewHMACCookieSigner([]byte("ffffffffffffffffffffffffffffffff"))
		_, err := other.Verify("session", signed)
		assert.ErrorIs(t, err, hime.ErrInvalidCookie)
	})

	t.Run("name swap fails", func(t *testing.T) {
		t.Parallel()
		signed, _ := signer.Sign("session", "user42")
		_, err := signer.Verify("other", signed) // name bound into the MAC
		assert.ErrorIs(t, err, hime.ErrInvalidCookie)
	})

	t.Run("malformed fails", func(t *testing.T) {
		t.Parallel()
		for _, in := range []string{"", "nodot", "!!!.QQ", "QQ.!!!", "."} {
			_, err := signer.Verify("session", in)
			assert.ErrorIs(t, err, hime.ErrInvalidCookie, "input %q", in)
		}
	})

	t.Run("empty key panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() { hime.NewHMACCookieSigner(nil) })
		assert.Panics(t, func() { hime.NewHMACCookieSigner([]byte{}) })
	})

	t.Run("key is copied", func(t *testing.T) {
		t.Parallel()
		key := append([]byte(nil), testCookieKey...)
		s := hime.NewHMACCookieSigner(key)
		signed, _ := s.Sign("session", "user42")

		for i := range key { // mutate the caller's slice after construction
			key[i] = 0
		}

		got, err := s.Verify("session", signed)
		assert.NoError(t, err)
		assert.Equal(t, "user42", got)
	})
}

func TestContextSignedCookie(t *testing.T) {
	t.Parallel()

	// signer used to generate signed values in tests; the app under test gets
	// an independent signer with the same key, so the two are compatible.
	signer := hime.NewHMACCookieSigner(testCookieKey)
	newApp := func() *hime.App {
		app := hime.New()
		app.CookieSigner = hime.NewHMACCookieSigner(testCookieKey)
		return app
	}

	t.Run("round trip through request and response", func(t *testing.T) {
		t.Parallel()
		app := newApp()

		w := httptest.NewRecorder()
		ctx := hime.NewAppContext(app, w, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.NoError(t, ctx.AddSignedCookie("session", "user42", &hime.CookieOptions{Path: "/"}))

		c := readSetCookie(w, "session")
		if !assert.NotNil(t, c) {
			return
		}
		assert.NotEqual(t, "user42", c.Value) // stored value is signed

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(c)
		ctx2 := hime.NewAppContext(app, httptest.NewRecorder(), r)
		assert.Equal(t, "user42", ctx2.SignedCookieValue("session"))
	})

	t.Run("missing cookie returns empty", func(t *testing.T) {
		t.Parallel()
		ctx := hime.NewAppContext(newApp(), httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, "", ctx.SignedCookieValue("session"))
	})

	t.Run("tampered cookie returns empty", func(t *testing.T) {
		t.Parallel()
		signed, _ := signer.Sign("session", "user42")
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "session", Value: flipFirstByte(signed)})
		ctx := hime.NewAppContext(newApp(), httptest.NewRecorder(), r)
		assert.Equal(t, "", ctx.SignedCookieValue("session"))
	})

	t.Run("value moved to another cookie name returns empty", func(t *testing.T) {
		t.Parallel()
		signed, _ := signer.Sign("session", "user42")
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "other", Value: signed})
		ctx := hime.NewAppContext(newApp(), httptest.NewRecorder(), r)
		assert.Equal(t, "", ctx.SignedCookieValue("other"))
	})

	t.Run("clone carries the signer", func(t *testing.T) {
		t.Parallel()
		clone := newApp().Clone()
		ctx := hime.NewAppContext(clone, httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		assert.NoError(t, ctx.AddSignedCookie("session", "x", nil)) // no panic => signer present
	})

	t.Run("AddSignedCookie propagates signer error", func(t *testing.T) {
		t.Parallel()
		app := hime.New()
		app.CookieSigner = failingSigner{}
		ctx := hime.NewAppContext(app, httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Error(t, ctx.AddSignedCookie("session", "user42", nil))
	})

	t.Run("no signer configured panics", func(t *testing.T) {
		t.Parallel()
		ctx := hime.NewAppContext(hime.New(), httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Panics(t, func() { ctx.AddSignedCookie("session", "x", nil) })
		assert.Panics(t, func() { ctx.SignedCookieValue("session") })
	})
}

// failingSigner is an external CookieSigner implementation used to confirm the
// interface is usable by callers and that Sign errors propagate.
type failingSigner struct{}

func (failingSigner) Sign(name, value string) (string, error) {
	return "", errors.New("sign failed")
}

func (failingSigner) Verify(name, signedValue string) (string, error) {
	return "", errors.New("verify failed")
}
