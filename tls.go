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
	"strings"
	"time"
)

// Restricted is the tls config for restricted mode
func Restricted() *tls.Config {
	return &tls.Config{
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
}

// Modern is the tls config for modern mode
func Modern() *tls.Config {
	return &tls.Config{
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
}

// Compatible is the tls config for compatible mode
func Compatible() *tls.Config {
	return &tls.Config{
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
}

// TLS type
type TLS struct {
	SelfSign   *SelfSign `yaml:"selfSign" json:"selfSign"`
	CertFile   string    `yaml:"certFile" json:"certFile"`
	KeyFile    string    `yaml:"keyFile" json:"keyFile"`
	Profile    string    `yaml:"profile" json:"profile"`
	MinVersion string    `yaml:"minVersion" json:"minVersion"`
	MaxVersion string    `yaml:"maxVersion" json:"maxVersion"`
	Curves     []string  `yaml:"curves" json:"curves"`
}

func parseTLSVersion(s string) uint16 {
	switch strings.ToLower(s) {
	case "":
		return 0
	case "ssl3.0":
		return tls.VersionSSL30
	case "tls1.0":
		return tls.VersionTLS10
	case "tls1.1":
		return tls.VersionTLS11
	case "tls1.2":
		return tls.VersionTLS12
	default:
		panicf("unknown tls version '%s'", s)
	}

	panic("unreachable")
}

func (t *TLS) config() *tls.Config {
	var tlsConfig *tls.Config

	switch strings.ToLower(t.Profile) {
	case "restricted":
		tlsConfig = Restricted()
	case "modern":
		tlsConfig = Modern()
	case "compatible":
		tlsConfig = Compatible()
	case "":
		tlsConfig = &tls.Config{}
	default:
		panicf("unknown tls profile '%s'", t.Profile)
	}

	tlsConfig.MinVersion = parseTLSVersion(t.MinVersion)
	tlsConfig.MaxVersion = parseTLSVersion(t.MaxVersion)

	if t.Curves != nil {
		tlsConfig.CurvePreferences = []tls.CurveID{}
		for _, c := range t.Curves {
			switch strings.ToLower(c) {
			case "p256":
				tlsConfig.CurvePreferences = append(tlsConfig.CurvePreferences, tls.CurveP256)
			case "p384":
				tlsConfig.CurvePreferences = append(tlsConfig.CurvePreferences, tls.CurveP384)
			case "p521":
				tlsConfig.CurvePreferences = append(tlsConfig.CurvePreferences, tls.CurveP521)
			case "x25519":
				tlsConfig.CurvePreferences = append(tlsConfig.CurvePreferences, tls.X25519)
			default:
				panicf("unknown tls curve '%s'", c)
			}
		}
	}

	if t.CertFile != "" && t.KeyFile != "" {
		err := loadTLSCertKey(tlsConfig, t.CertFile, t.KeyFile)
		if err != nil {
			panicf("load key pair error; %v", err)
		}
	} else if t.SelfSign != nil {
		t.SelfSign.config(tlsConfig)
	}

	return tlsConfig
}

// SelfSign type
type SelfSign struct {
	Key struct {
		Algo string `yaml:"algo" json:"algo"`
		Size int    `yaml:"size" json:"size"`
	} `yaml:"key" json:"key"`
	CN    string   `yaml:"cn" json:"cn"`
	Hosts []string `yaml:"host" json:"host"`
}

func loadTLSCertKey(t *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	t.Certificates = append(t.Certificates, cert)
	return nil
}

func (s *SelfSign) config(t *tls.Config) error {
	var priv interface{}
	var pub interface{}

	switch s.Key.Algo {
	case "ecdsa", "":
		var curve elliptic.Curve
		switch s.Key.Size {
		case 224:
			curve = elliptic.P224()
		case 256, 0:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		default:
			return fmt.Errorf("invalid self-sign key size '%d'", s.Key.Size)
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
		return fmt.Errorf("invalid self-sign key algo '%s'", s.Key.Algo)
	}

	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	cert := x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			CommonName:   s.CN,
			Organization: []string{"Acme Co"},
		},
		NotAfter:              time.Now().AddDate(1, 0, 0), // TODO: make not before configurable
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range s.Hosts {
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
