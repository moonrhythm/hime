package hime

// Handler is the hime handler
type Handler func(*Context) error

// Param is the query param when redirect
type Param struct {
	Name  string
	Value interface{}
}
