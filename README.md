# Hime

[![Test](https://github.com/moonrhythm/hime/actions/workflows/test.yaml/badge.svg)](https://github.com/moonrhythm/hime/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/moonrhythm/hime)](https://goreportcard.com/report/github.com/moonrhythm/hime)
[![GoDoc](https://godoc.org/github.com/moonrhythm/hime?status.svg)](https://godoc.org/github.com/moonrhythm/hime)

Hime is a Go Web Framework.

See [Wiki](https://github.com/moonrhythm/hime/wiki) for guide more information.

## Why Framework

I ❤️ net/http but... there are many duplicated code when working on multiple projects,
plus no standard. Framework creates a standard for developers.

### Why Another Framework

There is many Go frameworks out there. But I want a framework that works with any net/http compatible libraries seamlessly.

For example, you can choose any router, any middlewares, or handlers that work with standard library.

That why hime won't ship with any handler include router 🙈

## htmx

Hime ships opt-in [htmx](https://htmx.org) helpers so a server-rendered app can feel like a SPA — no client build step, no runtime beyond htmx itself.

One view, full page or fragment, from one handler:

```go
func settings(ctx *hime.Context) error {
	// full page for browsers, deep links, back button;
	// just the "content" block for htmx partial requests
	return ctx.ViewPartial("settings", "content", data)
}
```

```html
{{define "root"}}
<body hx-boost="true">
	<nav>...</nav>
	{{block "content" .}}...{{end}}
</body>
{{end}}
```

Notes:

- `ViewPartial` and htmx-aware redirects automatically add `Vary: HX-Request, HX-Boosted, HX-History-Restore-Request`, so caches never mix full pages and fragments. When you branch on `IsHTMX`/`IsBoosted`/`WantsPartial` yourself, call `ctx.VaryHTMX()` (chainable); branching on `HX-Target` too? Add `ctx.AddHeader("Vary", "HX-Target")` as well.
- Set `app.HTMXAwareRedirect = true` and the `Redirect` family answers htmx partial requests with `HX-Redirect` + 204 instead of a 3xx — post/redirect/get (with flash messages) works unchanged over htmx.
- `#` is reserved in template names: `ctx.View("page#form", data)` renders only that `{{block}}`/`{{define}}` of the view.
- Out-of-band swaps are a template pattern, not an API — put a conditional attribute on an id-carrying element inside the fragment you render (pass ctx as view data):

  ```html
  <div id="flash" {{if .Ctx.WantsPartial}}hx-swap-oob="true"{{end}}>
  	{{range index .Ctx.Flashes "success"}}<p>{{.}}</p>{{end}}
  </div>
  ```

- Active nav: keep navigation `hx-boost`-only (full pages re-render the nav and `ctx.IsRoute`), or make the nav an id-carrying block using the same conditional `hx-swap-oob` pattern.
- Form validation re-renders: respond `ctx.Status(422).ViewPartial(...)` with `FormState` errors, and allow htmx to swap 422 responses (the 422 rule must come before the `[45]..` rule — first match wins):

  ```html
  <meta name="htmx-config" content='{"responseHandling":[{"code":"422","swap":true},{"code":"204","swap":false},{"code":"[23]..","swap":true},{"code":"[45]..","swap":false,"error":true}]}'>
  ```

- CSRF: hime ships no CSRF middleware — wire your token with `hx-headers`, e.g. `<body hx-headers='{"X-CSRF-Token": "{{.Token}}"}'>`. Use per-session (not per-request) tokens, or set `hx-history="false"`, since htmx snapshots pages into localStorage.

## License

MIT
