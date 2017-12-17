package hime

// Route names a route
func (app *app) Route(name, path string) App {
	app.namedRoute[name] = path
	return app
}

func (app *app) GetRoute(name string) string {
	return app.namedRoute[name]
}
