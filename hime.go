package hime

import (
	"net/http"
)

// Routes is the map for route name => path
type Routes map[string]string

// Globals is the global const map
type Globals map[interface{}]interface{}

// Handler is the hime handler
type Handler func(*Context) Result

// Result is the handler result
type Result http.Handler

// Param is the query param when redirect
type Param struct {
	Name  string
	Value interface{}
}
