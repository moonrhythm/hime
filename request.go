package hime

import (
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func trimComma(s string) string {
	return strings.Replace(s, ",", "", -1)
}

func (ctx *appContext) ParseForm() error {
	return ctx.r.ParseForm()
}

func (ctx *appContext) ParseMultipartForm(maxMemory int64) error {
	return ctx.r.ParseMultipartForm(maxMemory)
}

func (ctx *appContext) Form() url.Values {
	return ctx.r.Form
}

func (ctx *appContext) PostForm() url.Values {
	return ctx.r.PostForm
}

func (ctx *appContext) FormValue(key string) string {
	return ctx.r.FormValue(key)
}

func (ctx *appContext) FormValueTrimSpace(key string) string {
	return strings.TrimSpace(ctx.FormValue(key))
}

func (ctx *appContext) FormValueTrimSpaceComma(key string) string {
	return trimComma(strings.TrimSpace(ctx.FormValue(key)))
}

func (ctx *appContext) FormValueInt(key string) int {
	x, _ := strconv.Atoi(ctx.FormValueTrimSpaceComma(key))
	return x
}

func (ctx *appContext) FormValueInt64(key string) int64 {
	x, _ := strconv.ParseInt(ctx.FormValueTrimSpaceComma(key), 10, 64)
	return x
}

func (ctx *appContext) FormValueFloat32(key string) float32 {
	x, _ := strconv.ParseFloat(ctx.FormValueTrimSpaceComma(key), 32)
	return float32(x)
}

func (ctx *appContext) FormValueFloat64(key string) float64 {
	x, _ := strconv.ParseFloat(ctx.FormValueTrimSpaceComma(key), 64)
	return float64(x)
}

func (ctx *appContext) PostFormValue(key string) string {
	return ctx.r.PostFormValue(key)
}

func (ctx *appContext) PostFormValueTrimSpace(key string) string {
	return strings.TrimSpace(ctx.PostFormValue(key))
}

func (ctx *appContext) PostFormValueTrimSpaceComma(key string) string {
	return trimComma(strings.TrimSpace(ctx.PostFormValue(key)))
}

func (ctx *appContext) PostFormValueInt(key string) int {
	x, _ := strconv.Atoi(ctx.PostFormValueTrimSpaceComma(key))
	return x
}

func (ctx *appContext) PostFormValueInt64(key string) int64 {
	x, _ := strconv.ParseInt(ctx.PostFormValueTrimSpaceComma(key), 10, 64)
	return x
}

func (ctx *appContext) PostFormValueFloat32(key string) float32 {
	x, _ := strconv.ParseFloat(ctx.PostFormValueTrimSpaceComma(key), 32)
	return float32(x)
}

func (ctx *appContext) PostFormValueFloat64(key string) float64 {
	x, _ := strconv.ParseFloat(ctx.PostFormValueTrimSpaceComma(key), 64)
	return float64(x)
}

func (ctx *appContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.r.FormFile(key)
}

func (ctx *appContext) FormFileNotEmpty(key string) (multipart.File, *multipart.FileHeader, error) {
	file, header, err := ctx.r.FormFile(key)
	if header.Size == 0 {
		return nil, nil, http.ErrMissingFile
	}
	return file, header, err
}

func (ctx *appContext) MultipartForm() *multipart.Form {
	return ctx.r.MultipartForm
}

func (ctx *appContext) MultipartReader() (*multipart.Reader, error) {
	return ctx.r.MultipartReader()
}

func (ctx *appContext) Method() string {
	return ctx.r.Method
}

func (ctx *appContext) Query() url.Values {
	return ctx.r.URL.Query()
}
