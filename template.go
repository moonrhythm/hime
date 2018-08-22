package hime

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"gopkg.in/yaml.v2"
)

// TemplateConfig is template config
type TemplateConfig struct {
	Dir        string              `yaml:"dir" json:"dir"`
	Root       string              `yaml:"root" json:"root"`
	Minify     bool                `yaml:"minify" json:"minify"`
	Components []string            `yaml:"components" json:"components"`
	List       map[string][]string `yaml:"list" json:"list"`
	Delims     []string            `yaml:"delims" json:"delims"`
}

// Template creates new template loader
func (app *App) Template() *Template {
	if app.template == nil {
		app.template = make(map[string]*tmpl)
	}
	return &Template{
		list: app.template,
		funcs: append([]template.FuncMap{template.FuncMap{
			"route":  app.Route,
			"global": app.Global,
		}}, app.templateFuncs...),
	}
}

// TemplateFuncs registers app's level template funcs
func (app *App) TemplateFuncs(funcs ...template.FuncMap) *App {
	app.templateFuncs = append(app.templateFuncs, funcs...)
	return app
}

// TemplateFunc registers an app's level template func
func (app *App) TemplateFunc(name string, f interface{}) *App {
	return app.TemplateFuncs(template.FuncMap{name: f})
}

type tmpl struct {
	template.Template
	m *minify.M
}

func (t *tmpl) Execute(w io.Writer, data interface{}) error {
	// t.m.Writer is too slow for short data (html)

	if t.m == nil {
		return t.Template.Execute(w, data)
	}

	buf := bytesPool.Get().(*bytes.Buffer)
	defer bytesPool.Put(buf)

	buf.Reset()
	err := t.Template.Execute(buf, data)
	if err != nil {
		return err
	}

	return t.m.Minify("text/html", w, buf)
}

// Template is template loader
type Template struct {
	list       map[string]*tmpl
	root       string
	dir        string
	leftDelim  string
	rightDelim string
	funcs      []template.FuncMap
	components []string
	minifier   *minify.M
}

// Config loads template config
func (tp *Template) Config(cfg TemplateConfig) *Template {
	tp.Dir(cfg.Dir)
	tp.Root(cfg.Root)
	if len(cfg.Delims) == 2 {
		tp.Delims(cfg.Delims[0], cfg.Delims[1])
	}
	tp.Component(cfg.Components...)
	for name, filenames := range cfg.List {
		tp.ParseFiles(name, filenames...)
	}
	if cfg.Minify {
		tp.Minify()
	}

	return tp
}

// ParseConfig parses template config data
func (tp *Template) ParseConfig(data []byte) *Template {
	var config TemplateConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		panicf("can not parse template config; %v", err)
	}
	return tp.Config(config)
}

// ParseConfigFile parses template config from file
func (tp *Template) ParseConfigFile(filename string) *Template {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panicf("read template config file; %v", err)
	}
	return tp.ParseConfig(data)
}

// Minify enables minify when render html, css, js
func (tp *Template) Minify() *Template {
	tp.minifier = minify.New()
	tp.minifier.AddFunc("text/html", html.Minify)
	tp.minifier.AddFunc("text/css", css.Minify)
	tp.minifier.AddFunc("text/javascript", js.Minify)

	// sets minify for parsed templates
	for _, t := range tp.list {
		t.m = tp.minifier
	}

	return tp
}

// Delims sets left and right delims
func (tp *Template) Delims(left, right string) *Template {
	tp.leftDelim = left
	tp.rightDelim = right
	return tp
}

// Root calls t.Lookup(name) after load template,
// empty string won't trigger t.Lookup
//
// default is ""
func (tp *Template) Root(name string) *Template {
	tp.root = name
	return tp
}

// Dir sets root directory when load template
//
// default is ""
func (tp *Template) Dir(path string) *Template {
	tp.dir = path
	return tp
}

// Funcs adds template funcs while load template
func (tp *Template) Funcs(funcs ...template.FuncMap) *Template {
	tp.funcs = append(tp.funcs, funcs...)
	return tp
}

// Func adds a template func while load template
func (tp *Template) Func(name string, f interface{}) *Template {
	return tp.Funcs(template.FuncMap{name: f})
}

// Component adds given templates to every templates
func (tp *Template) Component(filename ...string) *Template {
	tp.components = append(tp.components, filename...)
	return tp
}

func (tp *Template) newTemplate(name string, parser func(t *template.Template) *template.Template) {
	if _, ok := tp.list[name]; ok {
		panic(newErrTemplateDuplicate(name))
	}

	t := template.New("").
		Delims(tp.leftDelim, tp.rightDelim).
		Funcs(template.FuncMap{
			"templateName": func() string { return name },
			"param": func(name string, value interface{}) *Param {
				return &Param{Name: name, Value: value}
			},
		})

	// register funcs
	for _, fn := range tp.funcs {
		t.Funcs(fn)
	}

	t = parser(t)

	// parse components
	if len(tp.components) > 0 {
		t = template.Must(t.ParseFiles(joinTemplateDir(tp.dir, tp.components...)...))
	}

	if tp.root != "" {
		t = t.Lookup(tp.root)
	}

	if t == nil {
		panicf("no root layout")
	}

	tp.list[name] = &tmpl{
		Template: *t,
		m:        tp.minifier,
	}
}

// Parse parses template from text
func (tp *Template) Parse(name string, text string) *Template {
	tp.newTemplate(name, func(t *template.Template) *template.Template {
		return template.Must(t.New(name).Parse(text))
	})

	return tp
}

// ParseFiles loads template from file
func (tp *Template) ParseFiles(name string, filenames ...string) *Template {
	tp.newTemplate(name, func(t *template.Template) *template.Template {
		return template.Must(t.ParseFiles(joinTemplateDir(tp.dir, filenames...)...)).Lookup(filenames[0])
	})

	return tp
}

// ParseGlob loads template from pattern
func (tp *Template) ParseGlob(name string, pattern string) *Template {
	if tp.root == "" {
		panicf("parse glob can not use without root")
	}

	tp.newTemplate(name, func(t *template.Template) *template.Template {
		d := tp.dir
		if !strings.HasSuffix(d, "/") {
			d += "/"
		}
		return template.Must(t.ParseGlob(tp.dir + pattern))
	})

	return tp
}

func joinTemplateDir(dir string, filenames ...string) []string {
	xs := make([]string, len(filenames))
	for i, filename := range filenames {
		xs[i] = filepath.Join(dir, filename)
	}
	return xs
}

func cloneFuncMaps(xs []template.FuncMap) []template.FuncMap {
	if xs == nil {
		return nil
	}

	rs := make([]template.FuncMap, len(xs))
	for i := range xs {
		rs[i] = xs[i]
	}
	return rs
}

func cloneTmpl(xs map[string]*tmpl) map[string]*tmpl {
	if xs == nil {
		return nil
	}

	rs := make(map[string]*tmpl)
	for k, v := range xs {
		rs[k] = v
	}
	return rs
}
