package hime

import (
	"crypto/tls"
	"crypto/x509"
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

func TestParseTLSVersionSSL30AndCase(t *testing.T) {
	assert.Equal(t, uint16(tls.VersionSSL30), parseTLSVersion("ssl3.0"))
	assert.Equal(t, uint16(tls.VersionTLS12), parseTLSVersion("TLS1.2"))
	assert.Equal(t, uint16(tls.VersionTLS13), parseTLSVersion("Tls1.3"))
}

func TestTLSProfilesDetail(t *testing.T) {
	assert.Equal(t, uint16(tls.VersionTLS12), Restricted().MinVersion)
	assert.Equal(t, uint16(tls.VersionTLS12), Modern().MinVersion)
	assert.Equal(t, uint16(tls.VersionTLS10), Compatible().MinVersion)

	assert.Len(t, Restricted().CipherSuites, 6)
	assert.Len(t, Modern().CipherSuites, 10)
	assert.Len(t, Compatible().CipherSuites, 15)

	// only Compatible includes the legacy 3DES cipher
	assert.Contains(t, Compatible().CipherSuites, uint16(tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA))
	assert.NotContains(t, Restricted().CipherSuites, uint16(tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA))
}

func TestTLSConfigVersionOverride(t *testing.T) {
	c := (&TLS{Profile: "modern", MinVersion: "tls1.3", MaxVersion: "tls1.3"}).config()
	assert.Equal(t, uint16(tls.VersionTLS13), c.MinVersion)
	assert.Equal(t, uint16(tls.VersionTLS13), c.MaxVersion)
}

func TestTLSConfigCurvesOverrideProfile(t *testing.T) {
	c := (&TLS{Profile: "restricted", Curves: []string{"p384"}}).config()
	assert.Equal(t, []tls.CurveID{tls.CurveP384}, c.CurvePreferences)
}

func TestTLSConfigEmptyProfile(t *testing.T) {
	c := (&TLS{}).config()
	assert.NotNil(t, c)
	assert.Empty(t, c.CipherSuites)
	assert.Equal(t, uint16(0), c.MinVersion)
	assert.Equal(t, uint16(0), c.MaxVersion)
}

func TestSelfSignInvalidAlgo(t *testing.T) {
	opt := &SelfSign{}
	opt.Key.Algo = "invalid"
	assert.EqualError(t, opt.config(&tls.Config{}), "invalid self-sign key algo 'invalid'")
}

func TestSelfSignInvalidSizeMessage(t *testing.T) {
	opt := &SelfSign{}
	opt.Key.Algo = "ecdsa"
	opt.Key.Size = 999
	assert.EqualError(t, opt.config(&tls.Config{}), "invalid self-sign key size '999'")
}

func TestSelfSignCertContents(t *testing.T) {
	tc := tls.Config{}
	opt := &SelfSign{CN: "example.com", Hosts: []string{"10.0.0.1", "::1", "example.com"}}

	if assert.NoError(t, opt.config(&tc)) && assert.Len(t, tc.Certificates, 1) {
		cert, err := x509.ParseCertificate(tc.Certificates[0].Certificate[0])
		assert.NoError(t, err)
		assert.Equal(t, "example.com", cert.Subject.CommonName)
		assert.Len(t, cert.IPAddresses, 2)
		assert.Len(t, cert.DNSNames, 1)
		assert.Equal(t, "example.com", cert.DNSNames[0])
	}
}

func TestLoadTLSCertKey(t *testing.T) {
	c := &tls.Config{}
	assert.NoError(t, loadTLSCertKey(c, "testdata/server.crt", "testdata/server.key"))
	assert.Len(t, c.Certificates, 1)
}
