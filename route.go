package hime

func (app *app) Routes(routes Routes) App {
	for name, path := range routes {
		app.routes[name] = path
	}
	return app
}

func (app *app) Route(name string) string {
	path, ok := app.routes[name]
	if !ok {
		panic("hime: route not found")
	}
	return path
}
