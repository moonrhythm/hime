package hime

import (
	"log"
	"net/http"
)

// Wrap wraps hime handler with http.Handler
func Wrap(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app, ok := r.Context().Value(ctxKeyApp).(*app)
		if !ok {
			log.Panicf("hime: handler not pass from app")
		}
		ctx := newContext(app, w, r)
		h(ctx).Response(w, r)
	})
}
