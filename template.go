package hime

import (
	"html/template"
	"path/filepath"
)

// TemplateRoot calls t.Lookup(name) after load template,
// empty string won't trigger t.Lookup
//
// default is ""
func (app *App) TemplateRoot(name string) *App {
	app.template.root = name
	return app
}

// TemplateDir sets root directory when load template
//
// default is ""
func (app *App) TemplateDir(path string) *App {
	app.template.dir = path
	return app
}

// TemplateFuncs adds template funcs while load template
func (app *App) TemplateFuncs(funcs ...template.FuncMap) *App {
	app.template.funcs = append(app.template.funcs, funcs...)
	return app
}

// Component adds given templates to every templates
func (app *App) Component(filename ...string) *App {
	app.template.components = append(app.template.components, filename...)
	return app
}

// Template loads template into memory
func (app *App) Template(name string, filename ...string) *App {
	if app.template.list == nil {
		app.template.list = make(map[string]*template.Template)
	}

	if _, ok := app.template.list[name]; ok {
		panic(newErrTemplateDuplicate(name))
	}

	t := template.New("")

	t.Funcs(template.FuncMap{
		"templateName": func() string { return name },
		"route":        app.Route,
		"global":       app.Global,
		"param":        makeParam,
	})

	// register funcs
	for _, fn := range app.template.funcs {
		t.Funcs(fn)
	}

	// load templates and components
	fn := make([]string, len(filename))
	copy(fn, filename)
	fn = append(fn, app.template.components...)

	t = template.Must(t.ParseFiles(joinTemplateDir(app.template.dir, fn...)...))

	if app.template.root != "" {
		t = t.Lookup(app.template.root)
	}

	app.template.list[name] = t

	return app
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
