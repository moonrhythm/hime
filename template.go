package hime

import (
	"bytes"
	"html/template"
	"io"
	"path/filepath"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

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
		}}, app.templateFunc...),
	}
}

// TemplateFuncs registers app's level template funcs
func (app *App) TemplateFuncs(funcs ...template.FuncMap) *App {
	app.templateFunc = append(app.templateFunc, funcs...)
	return app
}

type tmpl struct {
	template.Template
	m *minify.M
}

func (t *tmpl) Execute(wr io.Writer, data interface{}) error {
	// t.m.Writer is too slow for short data (html)

	if t.m == nil {
		return t.Template.Execute(wr, data)
	}

	// TODO: this can optimize using pool
	buf := bytes.Buffer{}
	err := t.Template.Execute(&buf, data)
	if err != nil {
		return err
	}
	return t.m.Minify("text/html", wr, &buf)
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

// Component adds given templates to every templates
func (tp *Template) Component(filename ...string) *Template {
	tp.components = append(tp.components, filename...)
	return tp
}

// Parse loads template into memory
func (tp *Template) Parse(name string, filenames ...string) *Template {
	if _, ok := tp.list[name]; ok {
		panic(newErrTemplateDuplicate(name))
	}

	t := template.New("").
		Delims(tp.leftDelim, tp.rightDelim).
		Funcs(template.FuncMap{
			"templateName": func() string { return name },
			"param":        makeParam,
		})

	// register funcs
	for _, fn := range tp.funcs {
		t.Funcs(fn)
	}

	// load templates and components
	fn := make([]string, len(filenames))
	copy(fn, filenames)
	fn = append(fn, tp.components...)

	t = template.Must(t.ParseFiles(joinTemplateDir(tp.dir, fn...)...))

	if tp.root != "" {
		t = t.Lookup(tp.root)
	}

	tp.list[name] = &tmpl{
		Template: *t,
		m:        tp.minifier,
	}

	return tp
}

func makeParam(name string, value interface{}) *Param {
	return &Param{Name: name, Value: value}
}

func joinTemplateDir(dir string, filenames ...string) []string {
	xs := make([]string, len(filenames))
	for i, filename := range filenames {
		xs[i] = filepath.Join(dir, filename)
	}
	return xs
}
