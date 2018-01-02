package hime

import (
	"fmt"
	"path"
	"strings"
)

func buildPath(base string, params ...interface{}) string {
	xs := make([]string, len(params))
	for i, p := range params {
		xs[i] = fmt.Sprint(p)
	}
	if base == "" || (len(xs) > 0 && !strings.HasSuffix(base, "/")) {
		base += "/"
	}
	return base + path.Join(xs...)
}

func (app *app) Routes(routes Routes) App {
	for name, path := range routes {
		app.routes[name] = path
	}
	return app
}

func (app *app) Route(name string, params ...interface{}) string {
	path, ok := app.routes[name]
	if !ok {
		panic("hime: route not found")
	}
	return buildPath(path, params...)
}

func (ctx *appContext) Route(name string, params ...interface{}) string {
	return ctx.app.Route(name, params...)
}
