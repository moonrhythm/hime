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
			assert.Equal(t, "1", ctx.FormValueTrimSpace("a"))
			assert.Equal(t, "a b", ctx.FormValueTrimSpace("b"))
			assert.Equal(t, "1,234", ctx.FormValueTrimSpace("c"))
		})

		t.Run("TrimSpaceComma", func(t *testing.T) {
			assert.Equal(t, "1", ctx.FormValueTrimSpaceComma("a"))
			assert.Equal(t, "a b", ctx.FormValueTrimSpaceComma("b"))
			assert.Equal(t, "1234", ctx.FormValueTrimSpaceComma("c"))
		})

		t.Run("Int", func(t *testing.T) {
			assert.Equal(t, 1, ctx.FormValueInt("a"))
			assert.Equal(t, 0, ctx.FormValueInt("b"))
			assert.Equal(t, 1234, ctx.FormValueInt("c"))
		})

		t.Run("Int64", func(t *testing.T) {
			assert.Equal(t, int64(1), ctx.FormValueInt64("a"))
			assert.Equal(t, int64(0), ctx.FormValueInt64("b"))
			assert.Equal(t, int64(1234), ctx.FormValueInt64("c"))
		})

		t.Run("Float32", func(t *testing.T) {
			assert.Equal(t, float32(1), ctx.FormValueFloat32("a"))
			assert.Equal(t, float32(0), ctx.FormValueFloat32("b"))
			assert.Equal(t, float32(1234), ctx.FormValueFloat32("c"))
		})

		t.Run("Float64", func(t *testing.T) {
			assert.Equal(t, float64(1), ctx.FormValueFloat64("a"))
			assert.Equal(t, float64(0), ctx.FormValueFloat64("b"))
			assert.Equal(t, float64(1234), ctx.FormValueFloat64("c"))
		})
	})

	t.Run("PostFormValue", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/", bytes.NewBufferString("a=1&b=a%20b%20&c=1,234%20"))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := Context{Request: r}

		t.Run("TrimSpace", func(t *testing.T) {
			assert.Equal(t, "1", ctx.PostFormValueTrimSpace("a"))
			assert.Equal(t, "a b", ctx.PostFormValueTrimSpace("b"))
			assert.Equal(t, "1,234", ctx.PostFormValueTrimSpace("c"))
		})

		t.Run("TrimSpaceComma", func(t *testing.T) {
			assert.Equal(t, "1", ctx.PostFormValueTrimSpaceComma("a"))
			assert.Equal(t, "a b", ctx.PostFormValueTrimSpaceComma("b"))
			assert.Equal(t, "1234", ctx.PostFormValueTrimSpaceComma("c"))
		})

		t.Run("Int", func(t *testing.T) {
			assert.Equal(t, 1, ctx.PostFormValueInt("a"))
			assert.Equal(t, 0, ctx.PostFormValueInt("b"))
			assert.Equal(t, 1234, ctx.PostFormValueInt("c"))
		})

		t.Run("Int64", func(t *testing.T) {
			assert.Equal(t, int64(1), ctx.PostFormValueInt64("a"))
			assert.Equal(t, int64(0), ctx.PostFormValueInt64("b"))
			assert.Equal(t, int64(1234), ctx.PostFormValueInt64("c"))
		})

		t.Run("Float32", func(t *testing.T) {
			assert.Equal(t, float32(1), ctx.PostFormValueFloat32("a"))
			assert.Equal(t, float32(0), ctx.PostFormValueFloat32("b"))
			assert.Equal(t, float32(1234), ctx.PostFormValueFloat32("c"))
		})

		t.Run("Float64", func(t *testing.T) {
			assert.Equal(t, float64(1), ctx.PostFormValueFloat64("a"))
			assert.Equal(t, float64(0), ctx.PostFormValueFloat64("b"))
			assert.Equal(t, float64(1234), ctx.PostFormValueFloat64("c"))
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
			assert.Equal(t, http.ErrMissingFile, err)
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
			assert.Equal(t, http.ErrMissingFile, err)
			assert.Empty(t, h)
		})
	})
}
