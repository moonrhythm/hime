package hime

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

// Config is app's config
type Config struct {
	Globals  map[interface{}]interface{}
	Routes   map[string]string
	Template struct {
		Dir        string
		Root       string
		Components []string
		List       map[string][]string
	}
}

// Load loads config
//
// Example:
//
// globals:
//   data1: test
// routes:
//   index: /
//   about: /about
// template:
//   dir: view
//   root: layout
//   components:
//   - comp/comp1.tmpl
//   - comp/comp2.tmpl
//   list:
//     main.tmpl:
//     - main.tmpl
//     - _layout.tmpl
//     about.tmpl: [about.tmpl, _layout.tmpl]
func (app *App) Load(config Config) *App {
	app.Globals(config.Globals)
	app.Routes(config.Routes)
	app.templateDir = config.Template.Dir
	app.templateRoot = config.Template.Root
	app.Component(config.Template.Components...)

	for name, filenames := range config.Template.List {
		app.Template(name, filenames...)
	}

	return app
}

// LoadFromFile loads config from file
func (app *App) LoadFromFile(filename string) *App {
	fs, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	var config Config
	err = yaml.NewDecoder(fs).Decode(&config)
	if err != nil {
		panic(err)
	}

	app.Load(config)

	return app
}
