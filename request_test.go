package hime

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	t.Run("FormValue", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?a=1&b=a%20b%20&c=1,234%20", nil)
		ctx := Context{Request: r}

		t.Run("TrimSpace", func(t *testing.T) {
			assert.Equal(t, ctx.FormValueTrimSpace("a"), "1")
			assert.Equal(t, ctx.FormValueTrimSpace("b"), "a b")
			assert.Equal(t, ctx.FormValueTrimSpace("c"), "1,234")
		})

		t.Run("TrimSpaceComma", func(t *testing.T) {
			assert.Equal(t, ctx.FormValueTrimSpaceComma("a"), "1")
			assert.Equal(t, ctx.FormValueTrimSpaceComma("b"), "a b")
			assert.Equal(t, ctx.FormValueTrimSpaceComma("c"), "1234")
		})

		t.Run("Int", func(t *testing.T) {
			assert.Equal(t, ctx.FormValueInt("a"), 1)
			assert.Equal(t, ctx.FormValueInt("b"), 0)
			assert.Equal(t, ctx.FormValueInt("c"), 1234)
		})

		t.Run("Int64", func(t *testing.T) {
			assert.Equal(t, ctx.FormValueInt64("a"), int64(1))
			assert.Equal(t, ctx.FormValueInt64("b"), int64(0))
			assert.Equal(t, ctx.FormValueInt64("c"), int64(1234))
		})

		t.Run("Float32", func(t *testing.T) {
			assert.Equal(t, ctx.FormValueFloat32("a"), float32(1))
			assert.Equal(t, ctx.FormValueFloat32("b"), float32(0))
			assert.Equal(t, ctx.FormValueFloat32("c"), float32(1234))
		})

		t.Run("Float64", func(t *testing.T) {
			assert.Equal(t, ctx.FormValueFloat64("a"), float64(1))
			assert.Equal(t, ctx.FormValueFloat64("b"), float64(0))
			assert.Equal(t, ctx.FormValueFloat64("c"), float64(1234))
		})
	})

	t.Run("PostFormValue", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/", bytes.NewBufferString("a=1&b=a%20b%20&c=1,234%20"))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := Context{Request: r}

		t.Run("TrimSpace", func(t *testing.T) {
			assert.Equal(t, ctx.PostFormValueTrimSpace("a"), "1")
			assert.Equal(t, ctx.PostFormValueTrimSpace("b"), "a b")
			assert.Equal(t, ctx.PostFormValueTrimSpace("c"), "1,234")
		})

		t.Run("TrimSpaceComma", func(t *testing.T) {
			assert.Equal(t, ctx.PostFormValueTrimSpaceComma("a"), "1")
			assert.Equal(t, ctx.PostFormValueTrimSpaceComma("b"), "a b")
			assert.Equal(t, ctx.PostFormValueTrimSpaceComma("c"), "1234")
		})

		t.Run("Int", func(t *testing.T) {
			assert.Equal(t, ctx.PostFormValueInt("a"), 1)
			assert.Equal(t, ctx.PostFormValueInt("b"), 0)
			assert.Equal(t, ctx.PostFormValueInt("c"), 1234)
		})

		t.Run("Int64", func(t *testing.T) {
			assert.Equal(t, ctx.PostFormValueInt64("a"), int64(1))
			assert.Equal(t, ctx.PostFormValueInt64("b"), int64(0))
			assert.Equal(t, ctx.PostFormValueInt64("c"), int64(1234))
		})

		t.Run("Float32", func(t *testing.T) {
			assert.Equal(t, ctx.PostFormValueFloat32("a"), float32(1))
			assert.Equal(t, ctx.PostFormValueFloat32("b"), float32(0))
			assert.Equal(t, ctx.PostFormValueFloat32("c"), float32(1234))
		})

		t.Run("Float64", func(t *testing.T) {
			assert.Equal(t, ctx.PostFormValueFloat64("a"), float64(1))
			assert.Equal(t, ctx.PostFormValueFloat64("b"), float64(0))
			assert.Equal(t, ctx.PostFormValueFloat64("c"), float64(1234))
		})
	})

	t.Run("FormFile", func(t *testing.T) {
		b := bytes.Buffer{}
		w := multipart.NewWriter(&b)
		fw, _ := w.CreateFormFile("f1", "f1-name")
		fw.Write([]byte("some-data"))
		w.CreateFormFile("f2", "f2-name")
		w.Close()

		r := httptest.NewRequest("POST", "/", &b)
		r.Header.Set("Content-Type", w.FormDataContentType())
		ctx := Context{Request: r}

		t.Run("NotEmpty", func(t *testing.T) {
			f, h, err := ctx.FormFileNotEmpty("f1")
			assert.NoError(t, err)
			assert.NotEmpty(t, f)
			assert.NotEmpty(t, h)

			f, h, err = ctx.FormFileNotEmpty("f2")
			assert.Error(t, err)
			assert.Equal(t, err, http.ErrMissingFile)
			assert.Empty(t, f)
			assert.Empty(t, h)
		})

		t.Run("Header", func(t *testing.T) {
			h, err := ctx.FormFileHeader("f1")
			assert.NoError(t, err)
			assert.NotEmpty(t, h)

			// go1.11 bring back old behavior, that is empty file will return error,
			// like ctx.FormFileHeaderNotEmpty
			//
			// But in go1.10, it will return a header with size = 0
		})

		t.Run("HeaderNotEmpty", func(t *testing.T) {
			h, err := ctx.FormFileHeaderNotEmpty("f1")
			assert.NoError(t, err)
			assert.NotEmpty(t, h)

			h, err = ctx.FormFileHeaderNotEmpty("f2")
			assert.Error(t, err)
			assert.Equal(t, err, http.ErrMissingFile)
			assert.Empty(t, h)
		})
	})
}

func TestRequestFormValueEdgeCases(t *testing.T) {
	r := httptest.NewRequest("GET", "/?neg=-123&negc=-1,234&f=-3.14&sci=1e3&fstr=3.14&commas=,,&consec=1,,2", nil)
	ctx := Context{Request: r}

	assert.Equal(t, -123, ctx.FormValueInt("neg"))
	assert.Equal(t, -1234, ctx.FormValueInt("negc"))
	assert.Equal(t, int64(-1234), ctx.FormValueInt64("negc"))
	assert.Equal(t, -3.14, ctx.FormValueFloat64("f"))
	assert.Equal(t, float64(1000), ctx.FormValueFloat64("sci"))
	assert.Equal(t, 0, ctx.FormValueInt("fstr")) // strconv.Atoi fails on "3.14", returns 0
	assert.Equal(t, "", ctx.FormValueTrimSpaceComma("commas"))
	assert.Equal(t, "12", ctx.FormValueTrimSpaceComma("consec"))
}

func TestRequestFormFileMissingKey(t *testing.T) {
	b := bytes.Buffer{}
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("f1", "f1-name")
	fw.Write([]byte("data"))
	w.Close()

	r := httptest.NewRequest("POST", "/", &b)
	r.Header.Set("Content-Type", w.FormDataContentType())
	ctx := Context{Request: r}

	f, h, err := ctx.FormFileNotEmpty("missing")
	assert.Equal(t, http.ErrMissingFile, err)
	assert.Nil(t, f)
	assert.Nil(t, h)

	h2, err := ctx.FormFileHeader("missing")
	assert.Equal(t, http.ErrMissingFile, err)
	assert.Nil(t, h2)

	h3, err := ctx.FormFileHeaderNotEmpty("missing")
	assert.Equal(t, http.ErrMissingFile, err)
	assert.Nil(t, h3)
}

func TestRequestQueryValue(t *testing.T) {
	r := httptest.NewRequest("GET", "/?a=1&b=a%20b%20&c=1,234%20&neg=-12&f=-3.5", nil)
	ctx := Context{Request: r}

	assert.Equal(t, "1", ctx.QueryValueTrimSpace("a"))
	assert.Equal(t, "a b", ctx.QueryValueTrimSpace("b"))
	assert.Equal(t, "1,234", ctx.QueryValueTrimSpace("c"))
	assert.Equal(t, "1234", ctx.QueryValueTrimSpaceComma("c"))

	assert.Equal(t, 1, ctx.QueryValueInt("a"))
	assert.Equal(t, 1234, ctx.QueryValueInt("c"))
	assert.Equal(t, -12, ctx.QueryValueInt("neg"))
	assert.Equal(t, 0, ctx.QueryValueInt("b")) // not a number

	assert.Equal(t, int64(1234), ctx.QueryValueInt64("c"))
	assert.Equal(t, float32(1), ctx.QueryValueFloat32("a"))
	assert.Equal(t, float64(1234), ctx.QueryValueFloat64("c"))
	assert.Equal(t, -3.5, ctx.QueryValueFloat64("f"))
}

func TestRequestValueSlices(t *testing.T) {
	t.Run("FormValues and PostFormValues", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/?q=z", bytes.NewBufferString("a=1&a=2&b=3"))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := Context{Request: r}

		assert.Equal(t, []string{"1", "2"}, ctx.FormValues("a"))
		assert.Equal(t, []string{"1", "2"}, ctx.PostFormValues("a"))
		assert.Equal(t, []string{"3"}, ctx.PostFormValues("b"))
		// Form merges query + body; PostForm is body only
		assert.Equal(t, []string{"z"}, ctx.FormValues("q"))
		assert.Empty(t, ctx.PostFormValues("q"))
		assert.Empty(t, ctx.FormValues("missing"))
	})

	t.Run("QueryValues", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?a=1&a=2&b=3", nil)
		ctx := Context{Request: r}

		assert.Equal(t, []string{"1", "2"}, ctx.QueryValues("a"))
		assert.Equal(t, []string{"3"}, ctx.QueryValues("b"))
		assert.Empty(t, ctx.QueryValues("missing"))
	})

	t.Run("PostFormValues parses lazily", func(t *testing.T) {
		// PostFormValues is the first form access, so it must trigger parsing
		r := httptest.NewRequest("POST", "/", bytes.NewBufferString("x=1&x=2"))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := Context{Request: r}

		assert.Equal(t, []string{"1", "2"}, ctx.PostFormValues("x"))
	})
}

func TestRequestFormFileHeaderParseError(t *testing.T) {
	// A multipart Content-Type with a malformed body makes ParseMultipartForm
	// fail, and FormFileHeader propagates that error.
	r := httptest.NewRequest("POST", "/", strings.NewReader("not a valid multipart body"))
	r.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	ctx := Context{Request: r}

	h, err := ctx.FormFileHeader("f")
	assert.Error(t, err)
	assert.Nil(t, h)
}
