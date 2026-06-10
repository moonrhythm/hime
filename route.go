package hime

import (
	"context"
	"strings"
)

// Routes is the map for route name => path
type Routes map[string]string

func cloneRoutes(xs Routes) Routes {
	if xs == nil {
		return nil
	}
	rs := make(Routes)
	for k, v := range xs {
		rs[k] = v
	}
	return rs
}

// Routes registers route name and path
func (app *App) Routes(routes Routes) {
	if app.routes == nil {
		app.routes = make(Routes)
	}
	for name, path := range routes {
		app.routes[name] = path
	}
}

// Route gets route path from given name
func (app *App) Route(name string, params ...any) string {
	if app.routes == nil {
		panic(newErrRouteNotFound(name))
	}
	path, ok := app.routes[name]
	if !ok {
		panic(newErrRouteNotFound(name))
	}
	return buildPath(path, params...)
}

// Route gets route path from name
func (ctx *Context) Route(name string, params ...any) string {
	return ctx.app.Route(name, params...)
}

// IsRoute reports whether the named route is the most specific registered route
// matching the current request path: the path equals the route's path or is
// under it, and no other registered route matches more deeply. It is handy for
// highlighting the active navigation link — on /admin/users/42 the
// "/admin/users" route is active but its "/admin" parent is not. Pass ctx into
// the view data to use it from a template as {{.IsRoute "home"}}.
//
// Matching is path-segment aware ("/admin" matches "/admin/x" but not
// "/administrators"), so "/" is active only on exactly "/"; any query string or
// trailing slash in the route path is ignored. It panics if name is not a
// registered route, like Route.
func (ctx *Context) IsRoute(name string) bool {
	raw, ok := ctx.app.routes[name]
	if !ok {
		panic(newErrRouteNotFound(name))
	}

	cur := ctx.Request.URL.Path
	target := routePath(raw)
	if !pathUnder(cur, target) {
		return false
	}

	// not the active route if a longer registered route also matches
	for _, r := range ctx.app.routes {
		if p := routePath(r); len(p) > len(target) && pathUnder(cur, p) {
			return false
		}
	}
	return true
}

// routePath returns the path component of a route value: the query is stripped
// and any trailing slash removed, except for root "/".
func routePath(route string) string {
	p, _, _ := strings.Cut(route, "?")
	if p != "/" {
		p = strings.TrimRight(p, "/")
	}
	return p
}

// pathUnder reports whether path equals prefix or is a sub-path of it,
// comparing whole segments so "/admin" is under "/admin" and "/admin/x" but not
// "/administrators".
func pathUnder(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

// Route returns route value from context
func Route(ctx context.Context, name string, params ...any) string {
	return getApp(ctx).Route(name, params...)
}
