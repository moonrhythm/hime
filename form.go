package hime

import "net/url"

// FormState holds submitted form values together with per-field validation
// errors, so a form can be re-rendered with the user's input after a failed
// submission (instead of making them retype everything).
type FormState struct {
	values url.Values
	errors map[string][]string
}

// FormState returns a FormState seeded with the request's form values (query
// and body). Add validation errors with AddError, then pass it to a view to
// re-render the form.
func (ctx *Context) FormState() *FormState {
	if ctx.Request.Form == nil {
		ctx.Request.ParseMultipartForm(defaultMaxMemory)
	}

	// copy, so SetValue does not mutate the request's parsed form
	values := make(url.Values, len(ctx.Request.Form))
	for k, vs := range ctx.Request.Form {
		values[k] = append([]string(nil), vs...)
	}

	return &FormState{
		values: values,
		errors: map[string][]string{},
	}
}

// Value returns the first submitted value for name, or "".
func (fs *FormState) Value(name string) string {
	if vs := fs.values[name]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

// Values returns all submitted values for name (for multi-value fields such as
// checkboxes or multi-selects).
func (fs *FormState) Values(name string) []string {
	return fs.values[name]
}

// SetValue sets name's value, replacing any submitted value. Useful for
// pre-filling a form (such as an edit form) before rendering.
func (fs *FormState) SetValue(name, value string) {
	fs.values.Set(name, value)
}

// AddError records a validation error message for the field name. Calls
// accumulate.
func (fs *FormState) AddError(name, message string) {
	fs.errors[name] = append(fs.errors[name], message)
}

// Error returns the first error message for name, or "".
func (fs *FormState) Error(name string) string {
	if es := fs.errors[name]; len(es) > 0 {
		return es[0]
	}
	return ""
}

// Errors returns all error messages for name.
func (fs *FormState) Errors(name string) []string {
	return fs.errors[name]
}

// HasError reports whether name has any error.
func (fs *FormState) HasError(name string) bool {
	return len(fs.errors[name]) > 0
}

// HasErrors reports whether any field has an error.
func (fs *FormState) HasErrors() bool {
	return len(fs.errors) > 0
}
