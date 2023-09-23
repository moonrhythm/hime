package hime

import "context"

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

// Route returns route value from context
func Route(ctx context.Context, name string, params ...any) string {
	return getApp(ctx).Route(name, params...)
}
