package main

import (
	"log"
	"net/http"

	"github.com/acoshift/hime"
	"github.com/acoshift/middleware"
)

func main() {
	err := hime.New().
		Handler(routerFactory).
		ListenAndServe(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

type ctxKeyData struct{}

func routerFactory(app hime.App) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", hime.H(func(ctx hime.Context) hime.Result {
		return ctx.String(ctx.Value(ctxKeyData{}).(string))
	}))

	return middleware.Chain(
		injectData,
	)(mux)
}

func injectData(h http.Handler) http.Handler {
	return hime.H(func(ctx hime.Context) hime.Result {
		ctx.WithValue(ctxKeyData{}, "injected data!")
		return h
	})
}
