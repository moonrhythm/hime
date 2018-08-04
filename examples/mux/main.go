package main

import (
	"log"

	"github.com/acoshift/hime"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.Handle("/", hime.Handler(home))
	r.Handle("/about/{name}", hime.Handler(about))

	app := hime.New()
	app.Handler(r)
	app.Address(":8080")
	log.Fatal(app.ListenAndServe())
}

func home(ctx *hime.Context) error {
	return ctx.HTML([]byte(`
		<!doctype html>
		<h1>Home</h1>
		<a href="/about/1">About 1</a>
		<br>
		<a href="/about/2">About 2</a>
	`))
}

func about(ctx *hime.Context) error {
	vars := mux.Vars(ctx.Request())
	name := vars["name"]
	return ctx.HTML([]byte(`
		<!doctype html>
		<h1>About ` + name + `</h1>
		<a href="/">Home</a>
	`))
}
