package main

import (
	"net/http"
	"time"

	"github.com/acoshift/hime"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", hime.Wrap(func(ctx hime.Context) {
		ctx.View("index", map[string]interface{}{
			"Name": "Acoshift",
		})
	}))
	mux.Handle("/about", hime.Wrap(func(ctx hime.Context) {
		ctx.View("about", nil)
	}))

	hime.New().
		TemplateDir("view").
		TemplateRoot("layout").
		Component("_navbar.component.tmpl").
		Template("index", "index.tmpl", "_layout.tmpl").
		Template("about", "about.tmpl", "_layout.tmpl").
		Minify().
		Path("index", "/").
		Path("about", "/about").
		Router(mux).
		ShutdownTimeout(5 * time.Second).
		ListenAndServeGracefully(":8080")
}
