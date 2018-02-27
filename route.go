package hime

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
