package hime

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTLSVersion(t *testing.T) {
	t.Run("unknown value", func(t *testing.T) {
		assert.Panics(t, func() { parseTLSVersion("unknown") })
	})

	testCases := []struct {
		in  string
		out uint16
	}{
		{"", 0},
		{"ssl3.0", tls.VersionSSL30},
		{"tls1.0", tls.VersionTLS10},
		{"tls1.1", tls.VersionTLS11},
		{"tls1.2", tls.VersionTLS12},
	}
	for _, tC := range testCases {
		t.Run(fmt.Sprintf("parse %s", tC.in), func(t *testing.T) {
			assert.Equal(t, tC.out, parseTLSVersion(tC.in))
		})
	}
}

func TestTLSMode(t *testing.T) {
	assert.NotEmpty(t, Restricted())
	assert.NotEmpty(t, Modern())
	assert.NotEmpty(t, Compatible())
}

func TestCloneTLSNextProto(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, cloneTLSNextProto(nil))
	})

	t.Run("empty", func(t *testing.T) {
		p := cloneTLSNextProto(map[string]func(*http.Server, *tls.Conn, http.Handler){})

		assert.NotNil(t, p)
		assert.Empty(t, p)
	})

	t.Run("not empty", func(t *testing.T) {
		p := cloneTLSNextProto(map[string]func(*http.Server, *tls.Conn, http.Handler){
			"spdy/3": func(*http.Server, *tls.Conn, http.Handler) {},
		})

		assert.NotNil(t, p)
		assert.NotEmpty(t, p)
	})
}
