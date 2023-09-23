package hime

import (
	"html/template"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"gopkg.in/yaml.v3"
)

// TemplateConfig is template config
type TemplateConfig struct {
	Dir     string              `yaml:"dir" json:"dir"`
	Root    string              `yaml:"root" json:"root"`
	Minify  bool                `yaml:"minify" json:"minify"`
	Preload []string            `yaml:"preload" json:"preload"`
	List    map[string][]string `yaml:"list" json:"list"`
	Delims  []string            `yaml:"delims" json:"delims"`
}

// Template creates new template loader
func (app *App) Template() *Template {
	return &Template{
		parent:     template.Must(app.parent.Clone()),
		list:       app.template,
		components: app.component,
	}
}

// TemplateFuncs registers app's level template funcs
func (app *App) TemplateFuncs(funcs ...template.FuncMap) {
	for _, f := range funcs {
		app.parent.Funcs(f)
	}
}

// TemplateFunc registers an app's level template func
func (app *App) TemplateFunc(name string, f any) {
	app.TemplateFuncs(template.FuncMap{name: f})
}

type tmpl struct {
	*template.Template
	m *minify.M
}

func (t *tmpl) Execute(w io.Writer, data any) error {
	// t.m.Writer is too slow for short data (html)

	if t.m == nil {
		return t.Template.Execute(w, data)
	}

	buf := getBytes()
	defer putBytes(buf)

	err := t.Template.Execute(buf, data)
	if err != nil {
		return err
	}

	return t.m.Minify("text/html", w, buf)
}

// Template is template loader
type Template struct {
	parent     *template.Template
	list       map[string]*tmpl
	root       string
	fs         fs.FS
	dir        string
	components map[string]*tmpl
	minifier   *minify.M
	parsed     bool
}

// Config loads template config
func (tp *Template) Config(cfg TemplateConfig) {
	tp.Dir(cfg.Dir)
	tp.Root(cfg.Root)
	if len(cfg.Delims) == 2 {
		tp.Delims(cfg.Delims[0], cfg.Delims[1])
	}
	if cfg.Minify {
		tp.Minify()
	}
	tp.Preload(cfg.Preload...)
	for name, filenames := range cfg.List {
		tp.ParseFiles(name, filenames...)
	}
}

// ParseConfig parses template config data
func (tp *Template) ParseConfig(data []byte) {
	var config TemplateConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		panicf("can not parse template config; %v", err)
	}
	tp.Config(config)
}

// ParseConfigFile parses template config from file
func (tp *Template) ParseConfigFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panicf("read template config file; %v", err)
	}
	tp.ParseConfig(data)
}

type TemplateMinifyConfig struct {
	HTML minify.Minifier
	CSS  minify.Minifier
	JS   minify.Minifier
}

// MinifyWith enables minify with custom options, must call before parse
func (tp *Template) MinifyWith(cfg TemplateMinifyConfig) {
	tp.minifier = minify.New()
	if cfg.HTML != nil {
		tp.minifier.Add("text/html", cfg.HTML)
	}
	if cfg.CSS != nil {
		tp.minifier.Add("text/css", cfg.CSS)
	}
	if cfg.JS != nil {
		tp.minifier.Add("application/javascript", cfg.JS)
	}
}

// Minify enables minify when render html, css, js, must call before parse
func (tp *Template) Minify() {
	tp.MinifyWith(TemplateMinifyConfig{
		HTML: &html.Minifier{},
		CSS:  &css.Minifier{},
		JS:   &js.Minifier{},
	})
}

// Delims sets left and right delims
func (tp *Template) Delims(left, right string) {
	tp.parent.Delims(left, right)
}

// Root calls t.Lookup(name) after load template,
// empty string won't trigger t.Lookup
//
// default is ""
func (tp *Template) Root(name string) {
	tp.root = name
}

// Dir sets root directory when load template
//
// default is ""
func (tp *Template) Dir(path string) {
	tp.dir = path
}

// FS uses fs when load template
func (tp *Template) FS(fs fs.FS) {
	tp.fs = fs
}

// Funcs adds template funcs while load template
func (tp *Template) Funcs(funcs ...template.FuncMap) {
	for _, f := range funcs {
		tp.parent.Funcs(f)
	}
}

// Func adds a template func while load template
func (tp *Template) Func(name string, f any) {
	tp.Funcs(template.FuncMap{name: f})
}

// Preload loads given templates before every templates
func (tp *Template) Preload(filename ...string) {
	if tp.parsed {
		panicf("preload must call before parse")
	}
	if len(filename) == 0 {
		return
	}

	if tp.fs == nil {
		template.Must(tp.parent.ParseFiles(joinTemplateDir(tp.dir, filename...)...))
	} else {
		template.Must(tp.parent.ParseFS(tp.fs, joinTemplateDir(tp.dir, filename...)...))
	}
}

func (tp *Template) newTemplate(name string, parser func(t *template.Template) *template.Template) {
	if _, ok := tp.list[name]; ok {
		panic(newErrTemplateDuplicate(name))
	}

	t := template.Must(tp.parent.Clone()).
		Funcs(template.FuncMap{
			"templateName": func() string { return name },
		})

	t = parser(t)

	if tp.root != "" {
		t = t.Lookup(tp.root)
	}
	if t == nil {
		panicf("no root layout")
	}

	tp.list[name] = &tmpl{
		Template: t,
		m:        tp.minifier,
	}
	tp.parsed = true
}

func (tp *Template) newComponent(name string, parser func(t *template.Template) *template.Template) {
	if _, ok := tp.list[name]; ok {
		panic(newErrComponentDuplicate(name))
	}

	t := template.Must(tp.parent.Clone()).
		Funcs(template.FuncMap{
			"componentName": func() string { return name },
		})

	t = parser(t)

	if t == nil {
		panicf("nil component")
	}

	tp.components[name] = &tmpl{
		Template: t,
		m:        tp.minifier,
	}
	tp.parsed = true
}

// Parse parses template from text
func (tp *Template) Parse(name string, text string) {
	tp.newTemplate(name, func(t *template.Template) *template.Template {
		return template.Must(t.New(name).Parse(text))
	})
}

// ParseFiles loads template from file
func (tp *Template) ParseFiles(name string, filenames ...string) {
	tp.newTemplate(name, func(t *template.Template) *template.Template {
		if tp.fs == nil {
			t = template.Must(t.ParseFiles(joinTemplateDir(tp.dir, filenames...)...))
		} else {
			t = template.Must(t.ParseFS(tp.fs, joinTemplateDir(tp.dir, filenames...)...))
		}
		if tp.root == "" {
			t = t.Lookup(filenames[0])
		}
		return t
	})
}

// ParseGlob loads template from pattern
func (tp *Template) ParseGlob(name string, pattern string) {
	if tp.root == "" {
		panicf("parse glob can not use without root")
	}

	tp.newTemplate(name, func(t *template.Template) *template.Template {
		d := tp.dir
		if !strings.HasSuffix(d, "/") {
			d += "/"
		}
		if tp.fs == nil {
			return template.Must(t.ParseGlob(d + pattern))
		} else {
			return template.Must(t.ParseFS(tp.fs, d+pattern))
		}
	})
}

// Component loads html/template into component list
func (tp *Template) Component(ts ...*template.Template) {
	for _, t := range ts {
		name := t.Name()
		if name == "" {
			panicf("can not load empty name component")
		}

		if _, ok := tp.components[name]; ok {
			panicf("component '%s' already exists", name)
		}

		tp.components[name] = &tmpl{
			Template: t,
			m:        tp.minifier,
		}
	}
}

// ParseComponent parses component from text
func (tp *Template) ParseComponent(name string, text string) {
	tp.newComponent(name, func(t *template.Template) *template.Template {
		return template.Must(t.New(name).Parse(text))
	})
}

// ParseComponentFile loads component from file
func (tp *Template) ParseComponentFile(name string, filename string) {
	tp.newComponent(name, func(t *template.Template) *template.Template {
		if tp.fs == nil {
			t = template.Must(t.ParseFiles(joinTemplateDir(tp.dir, filename)...))
		} else {
			t = template.Must(t.ParseFS(tp.fs, joinTemplateDir(tp.dir, filename)...))
		}
		t = t.Lookup(path.Base(filename))
		return t
	})
}

func (app *App) renderComponent(name string, args ...any) template.HTML {
	t := app.component[name]
	if t == nil {
		panicf("component '%s' not found", name)
	}

	var d any
	switch len(args) {
	case 0:
	case 1:
		d = args[0]
	default:
		panicf("wrong number of data args for component '%s' want 0-1 got %d", name, len(args))
	}

	buf := getBytes()
	defer putBytes(buf)

	err := t.Execute(buf, d)
	if err != nil {
		panicf("component '%s' execute error: %v", name, err)
	}

	return template.HTML(buf.String())
}

func joinTemplateDir(dir string, filenames ...string) []string {
	xs := make([]string, len(filenames))
	for i, filename := range filenames {
		xs[i] = path.Join(dir, filename)
	}
	return xs
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

func tfParam(name string, value any) *Param {
	return &Param{Name: name, Value: value}
}

func tfTemplateName() string {
	return ""
}
