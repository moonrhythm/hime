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
		{"tls1.0", tls.VersionTLS10},
		{"tls1.1", tls.VersionTLS11},
		{"tls1.2", tls.VersionTLS12},
		{"tls1.3", tls.VersionTLS13},
	}
	for _, tC := range testCases {
		t.Run(fmt.Sprintf("parse %s", tC.in), func(t *testing.T) {
			assert.Equal(t, parseTLSVersion(tC.in), tC.out)
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

func TestSelfSign(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		tc := tls.Config{}

		assert.NoError(t, (&SelfSign{}).config(&tc))
	})

	t.Run("ECDSA", func(t *testing.T) {
		tc := tls.Config{}
		opt := &SelfSign{}
		opt.Key.Algo = "ecdsa"
		assert.NoError(t, opt.config(&tc))
	})

	t.Run("ECDSA/224", func(t *testing.T) {
		tc := tls.Config{}
		opt := &SelfSign{}
		opt.Key.Algo = "ecdsa"
		opt.Key.Size = 224
		assert.NoError(t, opt.config(&tc))
	})

	t.Run("ECDSA/256", func(t *testing.T) {
		tc := tls.Config{}
		opt := &SelfSign{}
		opt.Key.Algo = "ecdsa"
		opt.Key.Size = 256
		assert.NoError(t, opt.config(&tc))
	})

	t.Run("ECDSA/384", func(t *testing.T) {
		tc := tls.Config{}
		opt := &SelfSign{}
		opt.Key.Algo = "ecdsa"
		opt.Key.Size = 384
		assert.NoError(t, opt.config(&tc))
	})

	t.Run("ECDSA/521", func(t *testing.T) {
		tc := tls.Config{}
		opt := &SelfSign{}
		opt.Key.Algo = "ecdsa"
		opt.Key.Size = 521
		assert.NoError(t, opt.config(&tc))
	})

	t.Run("ECDSA/111", func(t *testing.T) {
		tc := tls.Config{}
		opt := &SelfSign{}
		opt.Key.Algo = "ecdsa"
		opt.Key.Size = 111
		assert.Error(t, opt.config(&tc))
	})

	t.Run("RSA", func(t *testing.T) {
		tc := tls.Config{}
		opt := &SelfSign{}
		opt.Key.Algo = "rsa"
		assert.NoError(t, opt.config(&tc))
	})

	t.Run("Host", func(t *testing.T) {
		tc := tls.Config{}
		opt := &SelfSign{}
		opt.Hosts = []string{
			"192.168.0.1",
			"localhost",
		}
		assert.NoError(t, opt.config(&tc))
	})
}

func TestTLSConfig(t *testing.T) {
	t.Run("Load crt, key file", func(t *testing.T) {
		tc := TLS{
			CertFile: "testdata/server.crt",
			KeyFile:  "testdata/server.key",
		}

		if assert.NotPanics(t, func() { tc.config() }) {
			assert.NotNil(t, tc.config())
		}
	})

	t.Run("Load invalid crt, key file", func(t *testing.T) {
		tc := TLS{
			CertFile: "testdata/server.key",
			KeyFile:  "testdata/server.crt",
		}

		assert.Panics(t, func() { tc.config() })
	})

	t.Run("Curves", func(t *testing.T) {
		tc := TLS{
			Curves: []string{"p256", "p384", "p521", "x25519"},
		}

		assert.NotPanics(t, func() { tc.config() })
	})

	t.Run("Curves invalid", func(t *testing.T) {
		tc := TLS{
			Curves: []string{"p"},
		}

		assert.Panics(t, func() { tc.config() })
	})

	t.Run("Profile", func(t *testing.T) {
		assert.NotPanics(t, func() { (&TLS{Profile: "restricted"}).config() })
		assert.NotPanics(t, func() { (&TLS{Profile: "RESTRICTED"}).config() })
		assert.NotPanics(t, func() { (&TLS{Profile: "modern"}).config() })
		assert.NotPanics(t, func() { (&TLS{Profile: "compatible"}).config() })
	})

	t.Run("Profile invalid", func(t *testing.T) {
		assert.Panics(t, func() { (&TLS{Profile: "super-good"}).config() })
	})
}
