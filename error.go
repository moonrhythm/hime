package hime

import (
	"errors"
	"fmt"
)

// Errors
var (
	ErrAppNotFound = errors.New("hime: app not found")
)

// ErrRouteNotFound is the error for route not found
type ErrRouteNotFound struct {
	Route string
}

func (err *ErrRouteNotFound) Error() string {
	return fmt.Sprintf("hime: route '%s' not found", err.Route)
}

func newErrRouteNotFound(route string) error {
	return &ErrRouteNotFound{route}
}

// ErrTemplateNotFound is the error for template not found
type ErrTemplateNotFound struct {
	Name string
}

func (err *ErrTemplateNotFound) Error() string {
	return fmt.Sprintf("hime: template '%s' not found", err.Name)
}

func newErrTemplateNotFound(name string) error {
	return &ErrTemplateNotFound{name}
}

// ErrTemplateDuplicate is the error for template duplicate
type ErrTemplateDuplicate struct {
	Name string
}

func (err *ErrTemplateDuplicate) Error() string {
	return fmt.Sprintf("hime: template '%s' already exists", err.Name)
}

func newErrTemplateDuplicate(name string) error {
	return &ErrTemplateDuplicate{name}
}
