package hime

import (
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig is hime app's config
type AppConfig struct {
	Globals   Globals          `yaml:"globals" json:"globals"`
	Routes    Routes           `yaml:"routes" json:"routes"`
	Templates []TemplateConfig `yaml:"templates" json:"templates"`
}

// Config merges config into app's config
//
// Example:
//
// globals:
//
//	data1: test
//
// routes:
//
//	index: /
//	about: /about
//
// templates:
//   - dir: view
//     root: layout
//     delims: ["{{", "}}"]
//     minify: true
//     preload:
//   - comp/comp1.tmpl
//   - comp/comp2.tmpl
//     list:
//     main.tmpl:
//   - main.tmpl
//   - _layout.tmpl
//     about.tmpl: [about.tmpl, _layout.tmpl]
func (app *App) Config(config AppConfig) {
	app.Globals(config.Globals)
	app.Routes(config.Routes)

	for _, cfg := range config.Templates {
		app.Template().Config(cfg)
	}
}

// ParseConfig parses config data
func (app *App) ParseConfig(data []byte) {
	var config AppConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		panicf("can not parse config; %v", err)
	}
	app.Config(config)
}

// ParseConfigFile parses config from file
func (app *App) ParseConfigFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panicf("can not read config from file; %v", err)
	}
	app.ParseConfig(data)
}
