package hime

import (
	"crypto/tls"
	"io/ioutil"
	"log"
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

			// TODO: auto generate self-signed tls if cert file, key file empty
			if t.CertFile != "" {
				app.certFile = t.CertFile
			}
			if t.KeyFile != "" {
				app.keyFile = t.KeyFile
			}

			app.srv.TLSConfig = tlsConfig
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
