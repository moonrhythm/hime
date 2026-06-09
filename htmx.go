package hime

import "encoding/json"

// htmx (https://htmx.org) response/request header helpers. They are thin
// wrappers over the HX-* headers, opt-in, and require no client runtime beyond
// htmx itself.

// IsHTMX reports whether the request was made by htmx, via the HX-Request
// header. Use it to render a partial/component for htmx and a full page
// otherwise.
func (ctx *Context) IsHTMX() bool {
	return ctx.Request.Header.Get("HX-Request") == "true"
}

// HTMXRedirect instructs htmx to perform a client-side redirect using the
// HX-Redirect response header. Prefer this over Redirect when responding to an
// htmx request, since htmx does not follow normal 3xx redirects. params are
// applied to url the same way as Redirect.
func (ctx *Context) HTMXRedirect(url string, params ...any) error {
	ctx.SetHeader("HX-Redirect", buildPath(url, params...))
	return nil
}

// HTMXRefresh instructs htmx to do a full page reload via the HX-Refresh header.
func (ctx *Context) HTMXRefresh() error {
	ctx.SetHeader("HX-Refresh", "true")
	return nil
}

// HTMXReswap overrides how htmx swaps the response (e.g. "outerHTML",
// "beforeend") via the HX-Reswap header. It returns ctx for chaining.
func (ctx *Context) HTMXReswap(strategy string) *Context {
	ctx.SetHeader("HX-Reswap", strategy)
	return ctx
}

// HTMXRetarget overrides the element htmx swaps the response into via the
// HX-Retarget header (a CSS selector). It returns ctx for chaining.
func (ctx *Context) HTMXRetarget(selector string) *Context {
	ctx.SetHeader("HX-Retarget", selector)
	return ctx
}

// HTMXTrigger triggers client-side events after the swap via the HX-Trigger
// header. With no detail it sends the bare event name; with one detail value it
// sends {event: detail} as JSON. It returns ctx for chaining, and panics if
// detail can not be marshalled or if more than one detail is given.
func (ctx *Context) HTMXTrigger(event string, detail ...any) *Context {
	switch len(detail) {
	case 0:
		ctx.SetHeader("HX-Trigger", event)
	case 1:
		b, err := json.Marshal(map[string]any{event: detail[0]})
		if err != nil {
			panicf("htmx trigger: %v", err)
		}
		ctx.SetHeader("HX-Trigger", string(b))
	default:
		panicf("htmx trigger: want 0-1 detail args, got %d", len(detail))
	}
	return ctx
}
