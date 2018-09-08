package main

import (
	"log"
	"net/http"

	"github.com/moonrhythm/hime"
	"github.com/acoshift/middleware"
)

func main() {
	err := hime.New().
		Handler(router()).
		Address(":8080").
		ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

type ctxKeyData struct{}

func router() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", hime.Handler(func(ctx *hime.Context) error {
		return ctx.String(ctx.Value(ctxKeyData{}).(string))
	}))

	return middleware.Chain(
		injectData,
	)(mux)
}

func injectData(h http.Handler) http.Handler {
	return hime.Handler(func(ctx *hime.Context) error {
		ctx.WithValue(ctxKeyData{}, "injected data!")
		return ctx.Handle(h)
	})
}
