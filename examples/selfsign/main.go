package main

import (
	"log"
	"net/http"

	"github.com/acoshift/hime"
)

var config = []byte(`
server:
  addr: :8080
  tls:
    selfSign: {}
`)

func main() {
	app := hime.New()
	app.ParseConfig(config)
	app.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	log.Fatal(app.ListenAndServe())
}
