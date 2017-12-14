package hime

// Path names a path
func (app *App) Path(name, path string) *App {
	app.namedPath[name] = path
	return app
}
