package hime

import (
	"fmt"
	"path"
)

func buildPath(base string, params ...interface{}) string {
	xs := make([]string, len(params)+1)
	xs[0] = base
	for i, p := range params {
		xs[i+1] = fmt.Sprint(p)
	}
	return path.Join(xs...)
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
