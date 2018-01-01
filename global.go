package hime

func (app *app) Globals(globals Globals) App {
	for key, value := range globals {
		app.globals[key] = value
	}
	return app
}

func (app *app) Global(key interface{}) interface{} {
	return app.globals[key]
}

func (ctx *appContext) Global(key interface{}) interface{} {
	return ctx.app.Global(key)
}
