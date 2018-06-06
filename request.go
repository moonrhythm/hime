package hime

import (
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	// defaultMaxMemory is http.defaultMaxMemory
	defaultMaxMemory = 32 << 20 // 32 MB
)

func removeComma(s string) string {
	return strings.Replace(s, ",", "", -1)
}

// ParseForm calls r.ParseForm
func (ctx *Context) ParseForm() error {
	return ctx.r.ParseForm()
}

// ParseMultipartForm calls r.ParseMultipartForm
func (ctx *Context) ParseMultipartForm(maxMemory int64) error {
	return ctx.r.ParseMultipartForm(maxMemory)
}

// Form calls r.Form
func (ctx *Context) Form() url.Values {
	return ctx.r.Form
}

// PostForm calls r.PostForm
func (ctx *Context) PostForm() url.Values {
	return ctx.r.PostForm
}

// FormValue calls r.FormValue
func (ctx *Context) FormValue(key string) string {
	return ctx.r.FormValue(key)
}

// FormValueTrimSpace trims space from form value
func (ctx *Context) FormValueTrimSpace(key string) string {
	return strings.TrimSpace(ctx.FormValue(key))
}

// FormValueTrimSpaceComma trims space and remove comma from form value
func (ctx *Context) FormValueTrimSpaceComma(key string) string {
	return removeComma(strings.TrimSpace(ctx.FormValue(key)))
}

// FormValueInt converts form value to int
func (ctx *Context) FormValueInt(key string) int {
	x, _ := strconv.Atoi(ctx.FormValueTrimSpaceComma(key))
	return x
}

// FormValueInt64 converts form value to int64
func (ctx *Context) FormValueInt64(key string) int64 {
	x, _ := strconv.ParseInt(ctx.FormValueTrimSpaceComma(key), 10, 64)
	return x
}

// FormValueFloat32 converts form value to float32
func (ctx *Context) FormValueFloat32(key string) float32 {
	x, _ := strconv.ParseFloat(ctx.FormValueTrimSpaceComma(key), 32)
	return float32(x)
}

// FormValueFloat64 converts form value to float64
func (ctx *Context) FormValueFloat64(key string) float64 {
	x, _ := strconv.ParseFloat(ctx.FormValueTrimSpaceComma(key), 64)
	return float64(x)
}

// PostFormValue calls r.PostFormValue
func (ctx *Context) PostFormValue(key string) string {
	return ctx.r.PostFormValue(key)
}

// PostFormValueTrimSpace trims space from post form value
func (ctx *Context) PostFormValueTrimSpace(key string) string {
	return strings.TrimSpace(ctx.PostFormValue(key))
}

// PostFormValueTrimSpaceComma trims space and remove comma from post form value
func (ctx *Context) PostFormValueTrimSpaceComma(key string) string {
	return removeComma(strings.TrimSpace(ctx.PostFormValue(key)))
}

// PostFormValueInt converts post form value to int
func (ctx *Context) PostFormValueInt(key string) int {
	x, _ := strconv.Atoi(ctx.PostFormValueTrimSpaceComma(key))
	return x
}

// PostFormValueInt64 converts post form value to int64
func (ctx *Context) PostFormValueInt64(key string) int64 {
	x, _ := strconv.ParseInt(ctx.PostFormValueTrimSpaceComma(key), 10, 64)
	return x
}

// PostFormValueFloat32 converts post form value to flost32
func (ctx *Context) PostFormValueFloat32(key string) float32 {
	x, _ := strconv.ParseFloat(ctx.PostFormValueTrimSpaceComma(key), 32)
	return float32(x)
}

// PostFormValueFloat64 converts post form value to flost64
func (ctx *Context) PostFormValueFloat64(key string) float64 {
	x, _ := strconv.ParseFloat(ctx.PostFormValueTrimSpaceComma(key), 64)
	return float64(x)
}

// FormFile returns r.FormFile
func (ctx *Context) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.r.FormFile(key)
}

// FormFileNotEmpty returns file from r.FormFile
// only when file size is not empty,
// or return http.ErrMissingFile if file is empty
func (ctx *Context) FormFileNotEmpty(key string) (multipart.File, *multipart.FileHeader, error) {
	file, header, err := ctx.r.FormFile(key)
	if err != nil {
		return nil, nil, err
	}
	if header.Size == 0 {
		file.Close()
		return nil, nil, http.ErrMissingFile
	}
	return file, header, err
}

// FormFileHeader returns file header for given key without open file
func (ctx *Context) FormFileHeader(key string) (*multipart.FileHeader, error) {
	// edit from http.Request.FormFile
	if ctx.r.MultipartForm == nil {
		err := ctx.r.ParseMultipartForm(defaultMaxMemory)
		if err != nil {
			return nil, err
		}
	}
	if ctx.r.MultipartForm != nil && ctx.r.MultipartForm.File != nil {
		if fhs := ctx.r.MultipartForm.File[key]; len(fhs) > 0 {
			return fhs[0], nil
		}
	}
	return nil, http.ErrMissingFile
}

// FormFileHeaderNotEmpty returns file header if not empty,
// or http.ErrMissingFile if file is empty
func (ctx *Context) FormFileHeaderNotEmpty(key string) (*multipart.FileHeader, error) {
	fh, err := ctx.FormFileHeader(key)
	if err != nil {
		return nil, err
	}
	if fh.Size == 0 {
		return nil, http.ErrMissingFile
	}
	return fh, nil
}

// MultipartForm returns r.MultipartForm
func (ctx *Context) MultipartForm() *multipart.Form {
	return ctx.r.MultipartForm
}

// MultipartReader returns r.MultipartReader
func (ctx *Context) MultipartReader() (*multipart.Reader, error) {
	return ctx.r.MultipartReader()
}

// Method return r.Method
func (ctx *Context) Method() string {
	return ctx.r.Method
}

// Query returns r.URL.Query
func (ctx *Context) Query() url.Values {
	return ctx.r.URL.Query()
}
