# Hime

[![Go Report Card](https://goreportcard.com/badge/github.com/acoshift/hime)](https://goreportcard.com/report/github.com/acoshift/hime)
[![GoDoc](https://godoc.org/github.com/acoshift/hime?status.svg)](https://godoc.org/github.com/acoshift/hime)

Hime is a Go Web Framework.

## Why Framework

I like net/http but... there are many duplicated code when working on multiple projects,
and no standard. Framework creates a standard which make developers to follow.

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

## What this framework DO NOT focus

- Speed
- One framework do everything

## License

MIT
