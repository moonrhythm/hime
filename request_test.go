package hime

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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
