package hime

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

// AppConfig is hime app's config
type AppConfig struct {
	Globals   Globals          `yaml:"globals" json:"globals"`
	Routes    Routes           `yaml:"routes" json:"routes"`
	Templates []TemplateConfig `yaml:"templates" json:"templates"`
	Server    struct {
		Addr              string            `yaml:"addr" json:"addr"`
		ReadTimeout       string            `yaml:"readTimeout" json:"readTimeout"`
		ReadHeaderTimeout string            `yaml:"readHeaderTimeout" json:"readHeaderTimeout"`
		WriteTimeout      string            `yaml:"writeTimeout" json:"writeTimeout"`
		IdleTimeout       string            `yaml:"idleTimeout" json:"idleTimeout"`
		ReusePort         *bool             `yaml:"reusePort" json:"reusePort"`
		TCPKeepAlive      string            `yaml:"tcpKeepAlive" json:"tcpKeepAlive"`
		ETag              *bool             `yaml:"eTag" json:"eTag"`
		H2C               *bool             `yaml:"h2c" json:"h2c"`
		GracefulShutdown  *GracefulShutdown `yaml:"gracefulShutdown" json:"gracefulShutdown"`
		TLS               *TLS              `yaml:"tls" json:"tls"`
		HTTPSRedirect     *HTTPSRedirect    `yaml:"httpsRedirect" json:"httpsRedirect"`
	} `yaml:"server" json:"server"`
}

func parseDuration(s string, t *time.Duration) {
	if s == "" {
		return
	}

	var err error
	*t, err = time.ParseDuration(s)
	if err != nil {
		panicf("can not parse duration; %v", err)
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
//   preload:
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
//   eTag: true
//   h2c: true
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
			app.srv.Addr = server.Addr
		}
		parseDuration(server.ReadTimeout, &app.srv.ReadTimeout)
		parseDuration(server.ReadHeaderTimeout, &app.srv.ReadHeaderTimeout)
		parseDuration(server.WriteTimeout, &app.srv.WriteTimeout)
		parseDuration(server.IdleTimeout, &app.srv.IdleTimeout)
		parseDuration(server.TCPKeepAlive, &app.tcpKeepAlive)

		if server.ReusePort != nil {
			app.reusePort = *server.ReusePort
		}
		if server.ETag != nil {
			app.ETag = *server.ETag
		}
		if server.H2C != nil {
			app.H2C = *server.H2C
		}

		if t := server.TLS; t != nil {
			app.srv.TLSConfig = server.TLS.config()
		}

		if server.GracefulShutdown != nil {
			app.gs = server.GracefulShutdown
		}

		if rd := server.HTTPSRedirect; rd != nil {
			go func() {
				err := StartHTTPSRedirectServer(rd.Addr)
				if err != nil {
					panicf("start https redirect server error; %v", err)
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
		panicf("can not parse config; %v", err)
	}
	return app.Config(config)
}

// ParseConfigFile parses config from file
func (app *App) ParseConfigFile(filename string) *App {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panicf("can not read config from file; %v", err)
	}
	return app.ParseConfig(data)
}
