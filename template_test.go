package hime

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	t.Run("ParseConfig", func(t *testing.T) {
		tp := New().Template()
		tp.ParseConfig([]byte(`
dir: testdata/template
root: l
minify: true
delims:
- "[["
- "]]"
components:
- a.tmpl
- b.tmpl
list:
  p:
  - p1.tmpl
  - p2.tmpl
  k:
  - k1.tmpl`))

		assert.Equal(t, "testdata/template", tp.dir)
		assert.Equal(t, "l", tp.root)
		assert.NotNil(t, tp.minifier)
		assert.Equal(t, "[[", tp.leftDelim)
		assert.Equal(t, "]]", tp.rightDelim)
		assert.Equal(t, []string{"a.tmpl", "b.tmpl"}, tp.components)
		assert.Contains(t, tp.list, "p")
		assert.Contains(t, tp.list, "k")
		assert.NotContains(t, tp.list, "a.tmpl")
		assert.NotContains(t, tp.list, "p1.tmpl")
	})

	t.Run("ParseConfig without root", func(t *testing.T) {
		tp := New().Template()
		tp.ParseConfig([]byte(`
dir: testdata/template
minify: true
delims:
- "[["
- "]]"
components:
- a.tmpl
- b.tmpl
list:
  p:
  - p1.tmpl
  - p2.tmpl
  k:
  - k1.tmpl`))

		assert.Equal(t, "testdata/template", tp.dir)
		assert.Empty(t, tp.root)
		assert.NotNil(t, tp.minifier)
		assert.Equal(t, "[[", tp.leftDelim)
		assert.Equal(t, "]]", tp.rightDelim)
		assert.Equal(t, []string{"a.tmpl", "b.tmpl"}, tp.components)
		assert.Contains(t, tp.list, "p")
		assert.Contains(t, tp.list, "k")
		assert.NotContains(t, tp.list, "a.tmpl")
		assert.NotContains(t, tp.list, "p1.tmpl")
	})

	t.Run("ParseConfig invalid", func(t *testing.T) {
		tp := New().Template()
		assert.Panics(t, func() { tp.ParseConfig([]byte(`invalidyamlbytes`)) })
	})

	t.Run("ParseConfigFile", func(t *testing.T) {
		tp := New().Template()
		tp.ParseConfigFile("testdata/template/config.yaml")

		assert.Equal(t, "testdata/template", tp.dir)
		assert.Equal(t, "l", tp.root)
		assert.NotNil(t, tp.minifier)
		assert.Equal(t, "[[", tp.leftDelim)
		assert.Equal(t, "]]", tp.rightDelim)
		assert.Equal(t, []string{"a.tmpl", "b.tmpl"}, tp.components)
		assert.Contains(t, tp.list, "p")
		assert.Contains(t, tp.list, "k")
		assert.NotContains(t, tp.list, "a.tmpl")
		assert.NotContains(t, tp.list, "p1.tmpl")
	})

	t.Run("ParseConfigFile not exists", func(t *testing.T) {
		tp := New().Template()
		assert.Panics(t, func() { tp.ParseConfigFile("testdata/template/not-exists-config.yaml") })
	})

	t.Run("Parse", func(t *testing.T) {
		tp := New().Template()
		tp.Parse("t", "Test Data")

		assert.Contains(t, tp.list, "t")
	})

	t.Run("Parse with component", func(t *testing.T) {
		tp := New().Template()
		tp.Dir("testdata/template")
		tp.Component("b.tmpl")
		tp.Parse("t", `Test Data {{template "b"}}`)

		if assert.Contains(t, tp.list, "t") {
			b := bytes.Buffer{}
			assert.NoError(t, tp.list["t"].Execute(&b, nil))
			assert.Equal(t, "Test Data b", b.String())
		}
	})

	t.Run("ParseFiles", func(t *testing.T) {
		tp := New().Template()
		tp.Dir("testdata/template")
		tp.Component("b.tmpl")
		tp.ParseFiles("t", "p1.tmpl")

		if assert.Contains(t, tp.list, "t") {
			b := bytes.Buffer{}
			assert.NoError(t, tp.list["t"].Execute(&b, nil))
			assert.Equal(t, "Test Data b", b.String())
		}
	})

	t.Run("ParseGlob", func(t *testing.T) {
		tp := New().Template()
		tp.Dir("testdata/template")
		tp.Root("b")
		tp.Component("b.tmpl")
		tp.ParseGlob("t", "*/**.tmpl")

		if assert.Contains(t, tp.list, "t") {
			b := bytes.Buffer{}
			assert.NoError(t, tp.list["t"].Execute(&b, nil))
			assert.Equal(t, "b", b.String())
		}
	})

	t.Run("ParseGlob without root", func(t *testing.T) {
		tp := New().Template()
		tp.Dir("testdata/template")
		tp.Component("b.tmpl")

		assert.Panics(t, func() { tp.ParseGlob("t", "*/**.tmpl") })
	})

	t.Run("Root not exists", func(t *testing.T) {
		tp := New().Template()
		tp.Root("root")
		assert.Panics(t, func() { tp.Parse("t", "Test Data") })
	})

	t.Run("Parse duplicate name", func(t *testing.T) {
		tp := New().Template()
		assert.NotPanics(t, func() { tp.Parse("t", "Test Data") })
		assert.Panics(t, func() { tp.Parse("t", "Test Data") })
	})

	t.Run("Minify before parse", func(t *testing.T) {
		tp := New().Template()
		tp.Minify()
		tp.Parse("t", "  <h1>  Test   </h1>")

		b := bytes.Buffer{}
		tp.list["t"].Execute(&b, nil)
		assert.Equal(t, "<h1>Test</h1>", b.String())
	})

	t.Run("Minify after parse", func(t *testing.T) {
		tp := New().Template()
		tp.Parse("t", "  <h1>  Test   </h1>")
		tp.Minify()

		b := bytes.Buffer{}
		tp.list["t"].Execute(&b, nil)
		assert.Equal(t, "<h1>Test</h1>", b.String())
	})

	t.Run("Minify execute error", func(t *testing.T) {
		tp := New().Template()
		tp.Minify()
		tp.Parse("t", "  <h1>  Test {{$.A.B}}   </h1>")

		b := bytes.Buffer{}
		assert.Error(t, tp.list["t"].Execute(&b, map[string]interface{}{"A": "a"}))
		assert.Empty(t, b.String())
	})

	t.Run("Func", func(t *testing.T) {
		tp := New().Template()
		tp.Func("n", func() string { return "abc" })
		tp.Parse("t", "{{ n }}")

		b := bytes.Buffer{}
		tp.list["t"].Execute(&b, nil)
		assert.Equal(t, "abc", b.String())
	})

	t.Run("Funcs", func(t *testing.T) {
		tp := New().Template()
		tp.Funcs(template.FuncMap{"n": func() string { return "abc" }})
		tp.Parse("t", "{{ n }}")

		b := bytes.Buffer{}
		tp.list["t"].Execute(&b, nil)
		assert.Equal(t, "abc", b.String())
	})

	t.Run("templateName", func(t *testing.T) {
		tp := New().Template()
		tp.Parse("t", "{{ templateName }}")

		b := bytes.Buffer{}
		tp.list["t"].Execute(&b, nil)
		assert.Equal(t, "t", b.String())
	})

	t.Run("param", func(t *testing.T) {
		app := New()
		app.Routes(Routes{"p": "/p"})
		tp := app.Template()
		tp.Parse("t", `<a href="{{route "p" (param "id" 1)}}">go</a>`)

		b := bytes.Buffer{}
		tp.list["t"].Execute(&b, nil)
		assert.Equal(t, `<a href="/p?id=1">go</a>`, b.String())
	})

	t.Run("cloneFuncMaps", func(t *testing.T) {
		assert.Nil(t, cloneFuncMaps(nil))
		assert.NotNil(t, cloneFuncMaps([]template.FuncMap{}))
		assert.Len(t, cloneFuncMaps([]template.FuncMap{template.FuncMap{"a": func() string { return "" }}}), 1)
	})

	t.Run("cloneTmpl", func(t *testing.T) {
		assert.Nil(t, cloneTmpl(nil))
		assert.NotNil(t, cloneTmpl(map[string]*tmpl{}))
		assert.Len(t, cloneTmpl(map[string]*tmpl{"a": &tmpl{}}), 1)
	})
}
