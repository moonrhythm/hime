package hime

import (
	"net/http/httptest"
	"testing"

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
