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

func panicf(format string, a ...any) {
	panic(fmt.Sprintf("hime: "+format, a...))
}

// ErrComponentNotFound is the error for component not found
type ErrComponentNotFound struct {
	Name string
}

func (err *ErrComponentNotFound) Error() string {
	return fmt.Sprintf("hime: component '%s' not found", err.Name)
}

func newErrComponentNotFound(name string) error {
	return &ErrComponentNotFound{name}
}

// ErrComponentDuplicate is the error for component duplicate
type ErrComponentDuplicate struct {
	Name string
}

func (err *ErrComponentDuplicate) Error() string {
	return fmt.Sprintf("hime: component '%s' already exists", err.Name)
}

func newErrComponentDuplicate(name string) error {
	return &ErrComponentDuplicate{name}
}
