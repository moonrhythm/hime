package hime

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

func buildPath(base string, params ...interface{}) string {
	xs := make([]string, 0, len(params))
	ps := make(url.Values)
	for _, p := range params {
		if v, ok := p.(url.Values); ok {
			for key, value := range v {
				for _, vv := range value {
					ps[key] = append(ps[key], vv)
				}
			}
			continue
		}
		xs = append(xs, strings.TrimPrefix(fmt.Sprint(p), "/"))
	}
	if base == "" || (len(xs) > 0 && !strings.HasSuffix(base, "/")) {
		base += "/"
	}
	qs := ps.Encode()
	if len(qs) > 0 {
		qs = "?" + qs
	}
	return base + path.Join(xs...) + qs
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
