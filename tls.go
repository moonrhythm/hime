package hime

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// Restricted is the tls config for restricted mode
var Restricted = tls.Config{
	MinVersion: tls.VersionTLS12,
	CurvePreferences: []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
	},
	PreferServerCipherSuites: true,
	CipherSuites: []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	},
}

// Modern is the tls config for modern mode
var Modern = tls.Config{
	MinVersion: tls.VersionTLS12,
	CurvePreferences: []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
	},
	PreferServerCipherSuites: true,
	CipherSuites: []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	},
}

// Compatible is the tls config for compatible mode
var Compatible = tls.Config{
	MinVersion: tls.VersionTLS10,
	CurvePreferences: []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
	},
	PreferServerCipherSuites: true,
	CipherSuites: []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	},
}

func loadTLSCertKey(t *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	t.Certificates = append(t.Certificates, cert)
	return nil
}

func generateSelfSign(t *tls.Config, algo string, size int, cn string, hosts []string) error {
	var priv interface{}
	var pub interface{}

	switch algo {
	case "ecdsa", "":
		var curve elliptic.Curve
		switch size {
		case 224:
			curve = elliptic.P224()
		case 256, 0:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		default:
			return fmt.Errorf("invalid self-sign key size '%d'", size)
		}

		pri, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return err
		}
		priv, pub = pri, &pri.PublicKey
	case "rsa":
		// TODO: make rsa key size configurable
		pri, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}
		priv, pub = pri, &pri.PublicKey
	default:
		return fmt.Errorf("invalid self-sign key algo '%s'", algo)
	}

	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	cert := x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			CommonName:   cn,
			Organization: []string{"Acme Co"},
		},
		NotAfter:              time.Now().AddDate(1, 0, 0), // TODO: make not before configurable
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		} else {
			cert.DNSNames = append(cert.DNSNames, h)
		}
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &cert, &cert, pub, priv)
	if err != nil {
		return err
	}

	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	var keyPemBlock pem.Block
	switch k := priv.(type) {
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return err
		}
		keyPemBlock = pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	case *rsa.PrivateKey:
		keyPemBlock = pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	}
	keyPem := pem.EncodeToMemory(&keyPemBlock)

	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return err
	}

	t.Certificates = append(t.Certificates, tlsCert)

	return nil
}
