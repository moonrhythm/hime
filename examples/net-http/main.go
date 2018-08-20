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
	app := hime.New()

	app.Template().
		Dir("view").
		Root("layout").
		Funcs(tmplFunc).
		Component("_navbar.component.tmpl").
		ParseFile("index", "index.tmpl", "_layout.tmpl").
		ParseFile("about", "about.tmpl", "_layout.tmpl").
		Minify()

	err := app.
		Routes(hime.Routes{
			"index":          "/",
			"about":          "/about",
			"api/json":       "/api/json",
			"api/json/error": "/api/json/error",
		}).
		Globals(hime.Globals{
			"github": "https://github.com/acoshift/hime",
		}).
		Handler(router(app)).
		Address(":8080").
		GracefulShutdown().
		Notify(func() {
			log.Println("Received terminate signal")
		}).
		Wait(5 * time.Second).
		Timeout(5 * time.Second).
		ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func router(app *hime.App) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(app.Route("index"), hime.Handler(indexHandler))
	mux.Handle(app.Route("about"), hime.Handler(aboutHandler))
	mux.Handle(app.Route("api/json"), hime.Handler(apiJSONHandler))
	mux.Handle(app.Route("api/json/error"), hime.Handler(apiJSONErrorHandler))
	return middleware.Chain(
		logRequestMethod,
		logRequestURI,
	)(mux)
}

func logRequestURI(h http.Handler) http.Handler {
	return hime.Handler(func(ctx *hime.Context) error {
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

func indexHandler(ctx *hime.Context) error {
	if ctx.Request().URL.Path != "/" {
		return ctx.RedirectTo("index")
	}

	return ctx.View("index", map[string]interface{}{
		"Name": "Acoshift",
	})
}

func aboutHandler(ctx *hime.Context) error {
	return ctx.View("about", nil)
}

func apiJSONHandler(ctx *hime.Context) error {
	return ctx.JSON(map[string]interface{}{
		"success": "ok",
	})
}

func apiJSONErrorHandler(ctx *hime.Context) error {
	return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{
		"error": "bad request",
	})
}
