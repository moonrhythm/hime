package main // import "github.com/moonrhythm/hime/example/selfsign"

import (
	"log"
	"net/http"

	"github.com/moonrhythm/hime"
)

func main() {
	app := hime.New()
	app.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	app.Address(":8080")
	app.SelfSign(hime.SelfSign{})

	log.Fatal(app.ListenAndServe())
}
