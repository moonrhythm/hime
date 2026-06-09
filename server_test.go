package hime

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPSRedirect(t *testing.T) {
	srv := HTTPSRedirect{}.Server()

	assert.NotNil(t, srv.Handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Host = "localhost"
	srv.Handler.ServeHTTP(w, r)
	assert.Equal(t, w.Header().Get("Location"), "https://localhost/test")
}

func TestHTTPSRedirectServerConfig(t *testing.T) {
	srv := HTTPSRedirect{Addr: ":9090"}.Server()

	assert.Equal(t, ":9090", srv.Addr)
	assert.Equal(t, 5*time.Second, srv.ReadTimeout)
	assert.Equal(t, 5*time.Second, srv.WriteTimeout)
	assert.NotNil(t, srv.Handler)
}

func TestHTTPSRedirectResponse(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?a=1&b=2", nil)
	r.Host = "example.com"

	HTTPSRedirect{}.Server().Handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "close", w.Header().Get("Connection"))
	assert.Equal(t, "https://example.com/test?a=1&b=2", w.Header().Get("Location"))
}

func TestStartHTTPSRedirectServerInvalidAddr(t *testing.T) {
	// An unparseable listen address makes ListenAndServe return immediately
	// with an error instead of blocking.
	assert.Error(t, StartHTTPSRedirectServer("bad:bad:bad"))
}
