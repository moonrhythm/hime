package hime

import (
	"net/http"
)

// Handler is the hime handler
type Handler func(*Context) Result

// Result is the handler result
type Result http.Handler

// Param is the query param when redirect
type Param struct {
	Name  string
	Value interface{}
}
