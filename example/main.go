package main

import (
	"net/http"
	"time"

	"github.com/acoshift/hime"
)

func main() {
	hime.New().
		TemplateDir("view").
		TemplateRoot("layout").
		Component("_navbar.component.tmpl").
		Template("index", "index.tmpl", "_layout.tmpl").
		Template("about", "about.tmpl", "_layout.tmpl").
		Minify().
		Path("index", "/").
		Path("about", "/about").
		Handler(routerFactory).
		GracefulShutdown().
		ShutdownTimeout(5 * time.Second).
		ListenAndServe(":8080")
}

func routerFactory(app hime.App) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(app.GetPath("index"), hime.Wrap(indexHandler))
	mux.Handle(app.GetPath("about"), hime.Wrap(aboutHandler))
	return mux
}

func indexHandler(ctx hime.Context) hime.Result {
	if ctx.Request().URL.Path != "/" {
		return ctx.RedirectTo("index")
	}
	return ctx.View("index", map[string]interface{}{
		"Name": "Acoshift",
	})
}

func aboutHandler(ctx hime.Context) hime.Result {
	return ctx.View("about", nil)
}
