package hime

import (
	"net/http"
)

// Wrap wraps hime handler with http.Handler
func Wrap(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)

		result := h(ctx)
		if result != nil {
			result.ServeHTTP(ctx.w, ctx.r)
		}
	})
}

// H is the short hand for Wrap
func H(h Handler) http.Handler {
	return Wrap(h)
}
