package hime

import (
	"bytes"
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
}
