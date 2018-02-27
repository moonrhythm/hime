package hime

import (
	"mime/multipart"
	"net/url"
)

// ParseForm runs Request.ParseForm
func (ctx *Context) ParseForm() error {
	return ctx.r.ParseForm()
}

// ParseMultipartForm runs Request.ParseMultipartForm
func (ctx *Context) ParseMultipartForm(maxMemory int64) error {
	return ctx.r.ParseMultipartForm(maxMemory)
}

// Form runs Request.Form
func (ctx *Context) Form() url.Values {
	return ctx.r.Form
}

// PostForm runs Request.PostForm
func (ctx *Context) PostForm() url.Values {
	return ctx.r.PostForm
}

// FormValue runs Request.FormValue
func (ctx *Context) FormValue(key string) string {
	return ctx.r.FormValue(key)
}

// PostFormValue runs Request.PostFormValue
func (ctx *Context) PostFormValue(key string) string {
	return ctx.r.PostFormValue(key)
}

// FormFile runs Request.FormFile
func (ctx *Context) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.r.FormFile(key)
}

// MultipartForm runs Request.MultipartForm
func (ctx *Context) MultipartForm() *multipart.Form {
	return ctx.r.MultipartForm
}

// MultipartReader runs Request.MultipartReader
func (ctx *Context) MultipartReader() (*multipart.Reader, error) {
	return ctx.r.MultipartReader()
}

// Method returns request method
func (ctx *Context) Method() string {
	return ctx.r.Method
}
