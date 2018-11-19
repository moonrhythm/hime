package hime

import "context"

// Globals is the global const map
type Globals map[interface{}]interface{}

// Globals registers global constants
func (app *App) Globals(globals Globals) *App {
	for key, value := range globals {
		app.globals.Store(key, value)
	}
	return app
}

// Global gets value from global storage
func (app *App) Global(key interface{}) interface{} {
	v, _ := app.globals.Load(key)
	return v
}

// Global returns global value
func (ctx *Context) Global(key interface{}) interface{} {
	return ctx.app.Global(key)
}

// Global returns global value from context
func Global(ctx context.Context, key interface{}) interface{} {
	return getApp(ctx).Global(key)
}
