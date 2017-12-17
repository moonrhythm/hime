package hime

// Route names a route
func (app *app) Route(name, path string) App {
	app.namedRoute[name] = path
	return app
}

func (app *app) GetRoute(name string) string {
	route, ok := app.namedRoute[name]
	if !ok {
		panic("hime: route not found")
	}
	return route
}
