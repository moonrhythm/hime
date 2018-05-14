package hime

// Globals registers global constants
func (app *App) Globals(globals Globals) *App {
	for key, value := range globals {
		app.globals[key] = value
	}
	return app
}

// Global gets value from global storage
func (app *App) Global(key interface{}) interface{} {
	return app.globals[key]
}

func (ctx *appContext) Global(key interface{}) interface{} {
	return ctx.app.Global(key)
}
