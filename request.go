package hime

import (
	"mime/multipart"
	"net/http"
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

// QueryValueTrimSpace trims space from query value
func (ctx *Context) QueryValueTrimSpace(key string) string {
	return strings.TrimSpace(ctx.Request.URL.Query().Get(key))
}

// QueryValueTrimSpaceComma trims space and remove comma from query value
func (ctx *Context) QueryValueTrimSpaceComma(key string) string {
	return removeComma(strings.TrimSpace(ctx.Request.URL.Query().Get(key)))
}

// QueryValueInt converts query value to int
func (ctx *Context) QueryValueInt(key string) int {
	x, _ := strconv.Atoi(ctx.QueryValueTrimSpaceComma(key))
	return x
}

// QueryValueInt64 converts query value to int64
func (ctx *Context) QueryValueInt64(key string) int64 {
	x, _ := strconv.ParseInt(ctx.QueryValueTrimSpaceComma(key), 10, 64)
	return x
}

// QueryValueFloat32 converts query value to float32
func (ctx *Context) QueryValueFloat32(key string) float32 {
	x, _ := strconv.ParseFloat(ctx.QueryValueTrimSpaceComma(key), 32)
	return float32(x)
}

// QueryValueFloat64 converts query value to float64
func (ctx *Context) QueryValueFloat64(key string) float64 {
	x, _ := strconv.ParseFloat(ctx.QueryValueTrimSpaceComma(key), 64)
	return float64(x)
}

// FormValues returns all form values associated with the given key,
// parsing the form first if necessary (query and body values)
func (ctx *Context) FormValues(key string) []string {
	if ctx.Request.Form == nil {
		ctx.Request.ParseMultipartForm(defaultMaxMemory)
	}
	return ctx.Request.Form[key]
}

// PostFormValues returns all post form values associated with the given key,
// parsing the form first if necessary (body values only)
func (ctx *Context) PostFormValues(key string) []string {
	if ctx.Request.PostForm == nil {
		ctx.Request.ParseMultipartForm(defaultMaxMemory)
	}
	return ctx.Request.PostForm[key]
}

// QueryValues returns all query string values associated with the given key
func (ctx *Context) QueryValues(key string) []string {
	return ctx.Request.URL.Query()[key]
}

// FormFileNotEmpty returns file from r.FormFile
// only when file size is not empty,
// or return http.ErrMissingFile if file is empty
func (ctx *Context) FormFileNotEmpty(key string) (multipart.File, *multipart.FileHeader, error) {
	file, header, err := ctx.FormFile(key)
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
	if ctx.MultipartForm == nil {
		err := ctx.ParseMultipartForm(defaultMaxMemory)
		if err != nil {
			return nil, err
		}
	}

	if ctx.MultipartForm != nil && ctx.MultipartForm.File != nil {
		if fhs := ctx.MultipartForm.File[key]; len(fhs) > 0 {
			return fhs[0], nil
		}
	}

	return nil, http.ErrMissingFile
}

// FormFileHeaderNotEmpty returns file header if not empty,
// or http.ErrMissingFile if file is empty
//
// This function will be deprecated after drop go1.10 support, since go1.11 bring back
// old behavior
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
