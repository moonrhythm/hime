package hime

// Path names a path
func (app *app) Path(name, path string) App {
	app.namedPath[name] = path
	return app
}

func (app *app) GetPath(name string) string {
	return app.namedPath[name]
}
