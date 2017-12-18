package hime

import (
	"mime/multipart"
	"net/url"
)

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

func (ctx *appContext) PostFormValue(key string) string {
	return ctx.r.PostFormValue(key)
}

func (ctx *appContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.r.FormFile(key)
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
