package hime

import (
	"encoding/json"
	"slices"
	"strings"
)

// htmx (https://htmx.org) response/request header helpers. They are thin
// wrappers over the HX-* headers, opt-in, and require no client runtime beyond
// htmx itself.

var htmxVaryHeaders = []string{
	"HX-Request",
	"HX-Boosted",
	"HX-History-Restore-Request",
}

// VaryHTMX appends the htmx request headers that affect response body selection
// (HX-Request, HX-Boosted, HX-History-Restore-Request) to Vary, deduplicated
// (case-insensitive) and after any pre-existing values, so caches never mix
// full pages, boosted pages, and fragments. Hime never calls it for you: call
// it whenever a response that branches on IsHTMX, IsBoosted, WantsPartial, or
// ViewPartial is cacheable (a CDN in front, or ETag revalidation). Responses
// served with Cache-Control: no-store or no-cache do not need it. It returns
// ctx for chaining.
func (ctx *Context) VaryHTMX() *Context {
	existing := make(map[string]bool)
	for _, v := range ctx.w.Header().Values("Vary") {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				existing[strings.ToLower(part)] = true
			}
		}
	}

	var toAdd []string
	for _, h := range htmxVaryHeaders {
		if existing[strings.ToLower(h)] {
			continue
		}
		toAdd = append(toAdd, h)
		existing[strings.ToLower(h)] = true
	}
	if len(toAdd) == 0 {
		return ctx
	}
	ctx.AddHeader("Vary", strings.Join(toAdd, ", "))
	return ctx
}

// IsHTMX reports whether the request was made by htmx, via the HX-Request
// header. When a handler branches on it, also call VaryHTMX so caches keep
// the variants apart.
func (ctx *Context) IsHTMX() bool {
	return ctx.Request.Header.Get("HX-Request") == "true"
}

// IsBoosted reports whether the request was made via htmx boost
// (HX-Boosted == "true"). When a handler branches on it, also call VaryHTMX
// so caches keep the variants apart.
func (ctx *Context) IsBoosted() bool {
	return ctx.Request.Header.Get("HX-Boosted") == "true"
}

// WantsPartial reports whether the handler may respond with a page fragment
// instead of a full document: an htmx request that is not boosted and not a
// history restore. Boosted and history-restore requests need full documents.
// ViewPartial branches on it for you; on cacheable responses, also call
// VaryHTMX so caches keep the variants apart.
func (ctx *Context) WantsPartial() bool {
	return ctx.IsHTMX() && !ctx.IsBoosted() &&
		ctx.Request.Header.Get("HX-History-Restore-Request") != "true"
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

// HTMXPushURL sets HX-Push-Url so htmx pushes url into the browser history.
// params are applied the same way as Redirect.
func (ctx *Context) HTMXPushURL(url string, params ...any) *Context {
	ctx.SetHeader("HX-Push-Url", buildPath(url, params...))
	return ctx
}

// HTMXReplaceURL sets HX-Replace-Url so htmx replaces the current history entry.
// params are applied the same way as Redirect.
func (ctx *Context) HTMXReplaceURL(url string, params ...any) *Context {
	ctx.SetHeader("HX-Replace-Url", buildPath(url, params...))
	return ctx
}

// HTMXNoPush sets HX-Push-Url to false so htmx does not update the browser URL.
func (ctx *Context) HTMXNoPush() *Context {
	ctx.SetHeader("HX-Push-Url", "false")
	return ctx
}

// HTMXReselect sets HX-Reselect so htmx selects a subset of the response via a
// CSS selector before swapping. It returns ctx for chaining.
func (ctx *Context) HTMXReselect(selector string) *Context {
	ctx.SetHeader("HX-Reselect", selector)
	return ctx
}

// HTMXLocation instructs htmx to perform a client-side soft navigation via
// HX-Location (path form only) and responds 204 No Content. params are applied
// the same way as Redirect.
func (ctx *Context) HTMXLocation(url string, params ...any) error {
	ctx.SetHeader("HX-Location", buildPath(url, params...))
	return ctx.NoContent()
}

// HTMXTrigger triggers client-side events after the swap via the HX-Trigger
// header. Repeated calls merge: bare event names when every entry has no
// detail, otherwise one JSON object. With no detail it records a bare event;
// with one detail value it records {event: detail}. It returns ctx for
// chaining, and panics if detail can not be marshalled or if more than one
// detail is given. The header is set immediately so it survives Redirect and
// NoContent (which bypass writeHeader). Merged events serialize in
// alphabetical (not call) order.
func (ctx *Context) HTMXTrigger(event string, detail ...any) *Context {
	return ctx.mergeHTMXTrigger(&ctx.htmxTrigger, "HX-Trigger", event, detail...)
}

// HTMXTriggerAfterSwap is like HTMXTrigger but sets HX-Trigger-After-Swap.
func (ctx *Context) HTMXTriggerAfterSwap(event string, detail ...any) *Context {
	return ctx.mergeHTMXTrigger(&ctx.htmxTriggerAfterSwap, "HX-Trigger-After-Swap", event, detail...)
}

// HTMXTriggerAfterSettle is like HTMXTrigger but sets HX-Trigger-After-Settle.
func (ctx *Context) HTMXTriggerAfterSettle(event string, detail ...any) *Context {
	return ctx.mergeHTMXTrigger(&ctx.htmxTriggerAfterSettle, "HX-Trigger-After-Settle", event, detail...)
}

func (ctx *Context) mergeHTMXTrigger(m *map[string]json.RawMessage, header, event string, detail ...any) *Context {
	if *m == nil {
		*m = make(map[string]json.RawMessage)
	}
	switch len(detail) {
	case 0:
		(*m)[event] = nil
	case 1:
		b, err := json.Marshal(detail[0])
		if err != nil {
			panicf("htmx trigger: %v", err)
		}
		(*m)[event] = b
	default:
		panicf("htmx trigger: want 0-1 detail args, got %d", len(detail))
	}

	allBare := true
	for _, v := range *m {
		if v != nil {
			allBare = false
			break
		}
	}
	if allBare {
		names := make([]string, 0, len(*m))
		for name := range *m {
			names = append(names, name)
		}
		slices.Sort(names)
		ctx.SetHeader(header, strings.Join(names, ", "))
		return ctx
	}

	b, err := json.Marshal(*m)
	if err != nil {
		panicf("htmx trigger: %v", err)
	}
	ctx.SetHeader(header, string(b))
	return ctx
}
