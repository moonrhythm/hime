package hime

import (
	"context"
	"net/http"
)

// Handler is the hime handler
type Handler func(*Context) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(NewContext(w, r))

	switch {
	case err == nil:
	case err == context.Canceled:
	default:
		panic(err)
	}
}
