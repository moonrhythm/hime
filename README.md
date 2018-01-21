# Hime

[![Go Report Card](https://goreportcard.com/badge/github.com/acoshift/hime)](https://goreportcard.com/report/github.com/acoshift/hime)
[![GoDoc](https://godoc.org/github.com/acoshift/hime?status.svg)](https://godoc.org/github.com/acoshift/hime)
[![Sourcegraph](https://sourcegraph.com/github.com/acoshift/hime/-/badge.svg)](https://sourcegraph.com/github.com/acoshift/hime?badge)

Hime is a Go Web Framework.

## Why Framework

I like net/http but... there are many duplicated code when working on multiple projects,
plus no standard. Framework creates a standard for developers.

### Why Another Framework

There're many Go framework out there. But I want a framework that works with any net/http compatible libraries seamlessly.

For example, you can choose any router, any middlewares, or handlers that work with standard library.

Other framework don't allow this. They have built-in router, framework-specific middlewares.

## Core focus

- Add standard to code
- Compatible with any net/http compatible router
- Compatible with http.Handler without code change
- Compatible with net/http middlewares without code change
- Use standard html/template for view
- Built-in core functions for build web server
- Reduce developer bug

## What is this framework DO NOT focus

- Speed
- One framework do everything

## Example

```go
func main() {
	hime.New().
		Template("index", "index.tmpl", "_layout.tmpl").
		Minify().
		Routes(hime.Routes{
			"index": "/",
		}).
		BeforeRender(addHeaderRender).
		Handler(routerFactory).
		GracefulShutdown().
		ListenAndServe(":8080")
}

func routerFactory(app hime.App) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(app.Route("index"), hime.H(indexHandler))
	return middleware.Chain(
		logRequestMethod,
		logRequestURI,
	)(mux)
}

func logRequestURI(h http.Handler) http.Handler {
	return hime.H(func(ctx hime.Context) hime.Result {
		log.Println(ctx.Request().RequestURI)
		return ctx.Handle(h)
	})
}

func logRequestMethod(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method)
		h.ServeHTTP(w, r)
	})
}

func addHeaderRender(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		h.ServeHTTP(w, r)
	})
}

func indexHandler(ctx hime.Context) hime.Result {
	if ctx.Request().URL.Path != "/" {
		return ctx.RedirectTo("index")
	}
	return ctx.View("index", map[string]interface{}{
		"Name": "Acoshift",
	})
}
```

## Compatibility with net/http

### Handler

Hime doesn't have built-in router, you can use any http.Handler.

`hime.Wrap` (or `hime.H` for short-hand) wraps hime.Handler into http.Handler, so you can use hime's handler anywhere in your router that support http.Handler.

### Middleware

Hime use native func(http.Handler) http.Handler for middleware.
You can use any middleware that compatible with this type.

```go
func logRequestMethod(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method)
		h.ServeHTTP(w, r)
	})
}
```

You can also use hime's handler with middleware

```go
func logRequestURI(h http.Handler) http.Handler {
	return hime.H(func(ctx hime.Context) hime.Result {
		log.Println(ctx.Request().RequestURI)
		return ctx.Handle(h)
	})
}
```

### Use hime.App as Handler

If you don't want to use hime's built-in graceful shutdown,
you can use hime.App as normal handler.

```go
func main() {
	app := hime.New().
		Template("index", "index.tmpl", "_layout.tmpl").
		Minify().
		Routes(hime.Routes{
			"index": "/",
		}).
		BeforeRender(addHeaderRender).
		Handler(routerFactory)

	http.ListenAndServe(":8080", app)
}
```

## Why return Result

Bacause many developers forgot to return to end handler

```go
func signInHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return // many return like this, sometime developers forgot about it
	}
	...
}
```

with Result

```go
func signInHandler(ctx hime.Context) hime.Result {
	username := r.FormValue("username")
	if username == "" {
		return ctx.Status(http.StatusBadRequest).Error("username required")
	}
	...
}
```

Why not return error, like this...

```go
func signInHandler(ctx hime.Context) error {
	username := r.FormValue("username")
	if username == "" {
		return ctx.Status(http.StatusBadRequest).Error("username required")
	}
	...
}
```

Then, what if you return an error ?

```go
return err
```

Hime won't handle error for you :D

You can see that hime won't response anything, you must handle on your own.

## Why some functions use panic

Hime try to reduce developer errors,
some error can detect while development.
Hime will panic for that type of errors.

## Useful handlers and middlewares

- [acoshift/middleware](https://github.com/acoshift/middleware)
- [acoshift/header](https://github.com/acoshift/header)
- [acoshift/session](https://github.com/acoshift/session)
- [acoshift/flash](https://github.com/acoshift/flash)
- [acoshift/webstatic](https://github.com/acoshift/webstatic)
- [acoshift/httprouter](https://github.com/acoshift/httprouter)
- [acoshift/redirecthttps](https://github.com/acoshift/redirecthttps)
- [acoshift/cors](https://github.com/acoshift/cors)
- [acoshift/hsts](https://github.com/acoshift/hsts)
- [acoshift/cachestatic](https://github.com/acoshift/cachestatic)
- [acoshift/gzip](https://github.com/acoshift/gzip)

## FAQ

- Will hime support automated let's encrypt ?
  > No, you can use hime as normal handler and use other lib to do this.

- Do you use hime in production ?
  > Yes, why not ? :D

## License

MIT
