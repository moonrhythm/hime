package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/acoshift/hime"
	"github.com/acoshift/middleware"
)

var tmplFunc = template.FuncMap{
	"toUpper": func(s string) string {
		return strings.ToUpper(s)
	},
}

func main() {
	hime.New().
		TemplateDir("view").
		TemplateRoot("layout").
		TemplateFuncs(tmplFunc).
		Component("_navbar.component.tmpl").
		Template("index", "index.tmpl", "_layout.tmpl").
		Template("about", "about.tmpl", "_layout.tmpl").
		Minify().
		Route("index", "/").
		Route("about", "/about").
		BeforeRender(addHeaderRender).
		Handler(routerFactory).
		GracefulShutdown().
		ShutdownTimeout(5 * time.Second).
		ListenAndServe(":8080")
}

func routerFactory(app hime.App) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(app.GetRoute("index"), hime.Wrap(indexHandler))
	mux.Handle(app.GetRoute("about"), hime.Wrap(aboutHandler))
	return middleware.Chain(
		logRequestMethod,
		logRequestURI,
	)(mux)
}

func logRequestURI(h http.Handler) http.Handler {
	return hime.Wrap(func(ctx hime.Context) hime.Result {
		log.Println(ctx.Request().RequestURI)
		return ctx.Handle(h)
	})
}

func logRequestMethod(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method)
		h.ServeHTTP(w, r)
	})
}

func addHeaderRender(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Custom-Header", "Hello")
		h.ServeHTTP(w, r)
	})
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
