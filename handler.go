package hime

import (
	"net/http"
)

// Handler is the hime handler
type Handler func(*Context) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)

	if err := h(ctx); err != nil {
		panic(err)
	}
}
