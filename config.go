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
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// AppConfig is hime app's config
type AppConfig struct {
	Globals   map[interface{}]interface{} `yaml:"globals" json:"globals"`
	Routes    map[string]string           `yaml:"routes" json:"routes"`
	Templates []TemplateConfig            `yaml:"templates" json:"templates"`
	Server    struct {
		Addr              string `yaml:"addr" json:"addr"`
		ReadTimeout       string `yaml:"readTimeout" json:"readTimeout"`
		ReadHeaderTimeout string `yaml:"readHeaderTimeout" json:"readHeaderTimeout"`
		WriteTimeout      string `yaml:"writeTimeout" json:"writeTimeout"`
		IdleTimeout       string `yaml:"idleTimeout" json:"idleTimeout"`
		GracefulShutdown  *struct {
			Timeout string `yaml:"timeout" json:"timeout"`
			Wait    string `yaml:"wait" json:"wait"`
		} `yaml:"gracefulShutdown" json:"gracefulShutdown"`
		TLS *struct {
			SelfSign *struct {
				Key struct {
					Algo string `yaml:"algo" json:"algo"`
					Size int    `yaml:"size" json:"size"`
				} `yaml:"key" json:"key"`
				CN    string   `yaml:"cn" json:"cn"`
				Hosts []string `yaml:"host" json:"host"`
			} `yaml:"selfSign" json:"selfSign"`
			CertFile   string   `yaml:"certFile" json:"certFile"`
			KeyFile    string   `yaml:"keyFile" json:"keyFile"`
			Profile    string   `yaml:"profile" json:"profile"`
			MinVersion string   `yaml:"minVersion" json:"minVersion"`
			MaxVersion string   `yaml:"maxVersion" json:"maxVersion"`
			Curves     []string `yaml:"curves" json:"curves"`
		} `yaml:"tls" json:"tls"`
		HTTPSRedirect *struct {
			Addr string `json:"addr"`
		} `yaml:"httpsRedirect" json:"httpsRedirect"`
	} `yaml:"server" json:"server"`
}

func parseDuration(s string, t *time.Duration) {
	if s == "" {
		return
	}
	var err error
	*t, err = time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
}

// Config merges config into app's config
//
// Example:
//
// globals:
//   data1: test
// routes:
//   index: /
//   about: /about
// templates:
// - dir: view
//   root: layout
//   delims: ["{{", "}}"]
//   minify: true
//   components:
//   - comp/comp1.tmpl
//   - comp/comp2.tmpl
//   list:
//     main.tmpl:
//     - main.tmpl
//     - _layout.tmpl
//     about.tmpl: [about.tmpl, _layout.tmpl]
// server:
//   readTimeout: 10s
//   readHeaderTimeout: 5s
//   writeTimeout: 5s
//   idleTimeout: 30s
//   gracefulShutdown:
//     timeout: 1m
//     wait: 5s
func (app *App) Config(config AppConfig) *App {
	app.Globals(config.Globals)
	app.Routes(config.Routes)

	for _, cfg := range config.Templates {
		app.Template().Config(cfg)
	}

	{
		// server config
		server := config.Server

		if server.Addr != "" {
			app.Addr = server.Addr
		}
		parseDuration(server.ReadTimeout, &app.ReadTimeout)
		parseDuration(server.ReadHeaderTimeout, &app.ReadHeaderTimeout)
		parseDuration(server.WriteTimeout, &app.WriteTimeout)
		parseDuration(server.IdleTimeout, &app.IdleTimeout)

		if t := server.TLS; t != nil {
			var tlsConfig *tls.Config

			switch strings.ToLower(t.Profile) {
			case "restricted":
				tlsConfig = Restricted.Clone()
			case "modern":
				tlsConfig = Modern.Clone()
			case "compatible":
				tlsConfig = Compatible.Clone()
			default:
				tlsConfig = &tls.Config{}
			}

			switch strings.ToLower(t.MinVersion) {
			case "ssl3.0":
				tlsConfig.MinVersion = tls.VersionSSL30
			case "tls1.0":
				tlsConfig.MinVersion = tls.VersionTLS10
			case "tls1.1":
				tlsConfig.MinVersion = tls.VersionTLS11
			case "tls1.2":
				tlsConfig.MinVersion = tls.VersionTLS12
			}

			switch strings.ToLower(t.MaxVersion) {
			case "ssl3.0":
				tlsConfig.MaxVersion = tls.VersionSSL30
			case "tls1.0":
				tlsConfig.MaxVersion = tls.VersionTLS10
			case "tls1.1":
				tlsConfig.MaxVersion = tls.VersionTLS11
			case "tls1.2":
				tlsConfig.MaxVersion = tls.VersionTLS12
			}

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
						log.Panicf("hime: unknown tls curve '%s'", c)
					}
				}
			}

			if t.CertFile != "" && t.KeyFile != "" {
				cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
				if err != nil {
					panic("hime: load key pair error; " + err.Error())
				}
				tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
			} else if c := t.SelfSign; c != nil {
				var priv interface{}
				var pub interface{}

				switch c.Key.Algo {
				case "ecdsa", "":
					var curve elliptic.Curve
					switch c.Key.Size {
					case 224:
						curve = elliptic.P224()
					case 256, 0:
						curve = elliptic.P256()
					case 384:
						curve = elliptic.P384()
					case 521:
						curve = elliptic.P521()
					default:
						panic("hime: invalid self-sign key size")
					}

					pri, err := ecdsa.GenerateKey(curve, rand.Reader)
					if err != nil {
						panic("hime: generate private key error;" + err.Error())
					}
					priv, pub = pri, &pri.PublicKey
				case "rsa":
					// TODO: make rsa key size configurable
					pri, err := rsa.GenerateKey(rand.Reader, 2048)
					if err != nil {
						panic("hime: generate private key error; " + err.Error())
					}
					priv, pub = pri, &pri.PublicKey
				default:
					panic("hime: invalid self-sign key algo")
				}

				sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
				if err != nil {
					panic("hime: generate serial number error; " + err.Error())
				}

				cert := x509.Certificate{
					SerialNumber: sn,
					Subject: pkix.Name{
						Organization: []string{"Acme Co"},
					},
					NotAfter:              time.Now().AddDate(1, 0, 0), // TODO: make not before configurable
					KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
					ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
					BasicConstraintsValid: true,
				}

				for _, h := range c.Hosts {
					if ip := net.ParseIP(h); ip != nil {
						cert.IPAddresses = append(cert.IPAddresses, ip)
					} else {
						cert.DNSNames = append(cert.DNSNames, h)
					}
				}

				certBytes, err := x509.CreateCertificate(rand.Reader, &cert, &cert, pub, priv)
				if err != nil {
					panic("hime: create cvertificate error; " + err.Error())
				}

				certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

				var keyPemBlock pem.Block
				switch k := priv.(type) {
				case *ecdsa.PrivateKey:
					b, err := x509.MarshalECPrivateKey(k)
					if err != nil {
						panic("hime: marshal ec private key error; " + err.Error())
					}
					keyPemBlock = pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
				case *rsa.PrivateKey:
					keyPemBlock = pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
				}
				keyPem := pem.EncodeToMemory(&keyPemBlock)

				tlsCert, err := tls.X509KeyPair(certPem, keyPem)
				if err != nil {
					panic("hime: load key pair error; " + err.Error())
				}
				tlsConfig.Certificates = append(tlsConfig.Certificates, tlsCert)
			}

			app.TLSConfig = tlsConfig
		}

		if gs := server.GracefulShutdown; gs != nil {
			if app.gracefulShutdown == nil {
				app.gracefulShutdown = &gracefulShutdown{}
			}

			parseDuration(gs.Timeout, &app.gracefulShutdown.timeout)
			parseDuration(gs.Wait, &app.gracefulShutdown.wait)
		}

		if rd := server.HTTPSRedirect; rd != nil {
			if rd.Addr == "" {
				rd.Addr = ":80"
			}

			go func() {
				err := StartHTTPSRedirectServer(rd.Addr)
				if err != nil {
					log.Panicf("hime: start https redirect server error; %v", err)
				}
			}()
		}
	}

	return app
}

// ParseConfig parses config data
func (app *App) ParseConfig(data []byte) *App {
	var config AppConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	return app.Config(config)
}

// ParseConfigFile parses config from file
func (app *App) ParseConfigFile(filename string) *App {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return app.ParseConfig(data)
}
