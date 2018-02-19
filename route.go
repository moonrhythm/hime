package hime

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

func mergeValues(s, p url.Values) {
	for k, v := range p {
		for _, vv := range v {
			s[k] = append(s[k], vv)
		}
	}
}

func mergeValueWithMapString(s url.Values, m map[string]string) {
	for k, v := range m {
		s[k] = append(s[k], v)
	}
}

func mergeValueWithMapInterface(s url.Values, m map[string]interface{}) {
	for k, v := range m {
		s[k] = append(s[k], fmt.Sprint(v))
	}
}

func buildPath(base string, params ...interface{}) string {
	xs := make([]string, 0, len(params))
	ps := make(url.Values)
	for _, p := range params {
		switch v := p.(type) {
		case url.Values:
			mergeValues(ps, v)
		case map[string]string:
			mergeValueWithMapString(ps, v)
		case map[string]interface{}:
			mergeValueWithMapInterface(ps, v)
		default:
			xs = append(xs, strings.TrimPrefix(fmt.Sprint(p), "/"))
		}
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
		panic(newErrRouteNotFound(name))
	}
	return buildPath(path, params...)
}

func (ctx *appContext) Route(name string, params ...interface{}) string {
	return ctx.app.Route(name, params...)
}
