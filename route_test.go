package hime

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute(t *testing.T) {
	t.Run("clone nil", func(t *testing.T) {
		assert.Nil(t, cloneRoutes(nil))
	})

	t.Run("clone empty", func(t *testing.T) {
		r := Routes{}
		assert.Equal(t, cloneRoutes(r), Routes{})
	})

	t.Run("clone data", func(t *testing.T) {
		r := Routes{
			"a": "/b",
			"b": "/cd",
		}

		p := cloneRoutes(r)
		assert.Equal(t, p, r)

		p["a"] = "/a"
		assert.NotEqual(t, p, r)
	})

	t.Run("App", func(t *testing.T) {
		t.Run("panic when retrieve route from empty route app", func(t *testing.T) {
			app := New()

			assert.Panics(t, func() { app.Route("r1") })
		})

		t.Run("retrieve valid route", func(t *testing.T) {
			app := New()
			app.Routes(Routes{
				"a": "/b",
				"b": "/cd",
			})

			assert.Equal(t, app.Route("a"), "/b")
			assert.Equal(t, app.Route("b"), "/cd")
		})
	})

	t.Run("panic when retrieve not exists route", func(t *testing.T) {
		app := New()
		app.Routes(Routes{
			"a": "/b",
			"b": "/cd",
		})

		assert.Panics(t, func() { app.Route("c") })
	})

	t.Run("Context", func(t *testing.T) {
		t.Run("retrieve route from context", func(t *testing.T) {
			app := New()
			app.Routes(Routes{
				"a": "/b",
				"b": "/cd",
			})

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			app.Handler(Handler(func(ctx *Context) error {
				assert.Equal(t, ctx.Route("a"), "/b")
				assert.Equal(t, ctx.Route("b"), "/cd")
				assert.Equal(t, Route(ctx, "a"), "/b")
				assert.Equal(t, Route(ctx, "b"), "/cd")
				return nil
			}))
			app.ServeHTTP(w, r)
		})

		t.Run("panic when retrieve not exists route", func(t *testing.T) {
			app := New()
			app.Routes(Routes{
				"a": "/b",
				"b": "/cd",
			})

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			app.Handler(Handler(func(ctx *Context) error {
				assert.Panics(t, func() { ctx.Route("c") })
				return nil
			}))
			app.ServeHTTP(w, r)
		})
	})
}

func TestRouteWithParams(t *testing.T) {
	app := New()
	app.Routes(Routes{"user": "/user"})

	assert.Equal(t, "/user?id=10", app.Route("user", url.Values{"id": []string{"10"}}))
	assert.Equal(t, "/user?id=10", app.Route("user", map[string]string{"id": "10"}))
	assert.Equal(t, "/user/10", app.Route("user", "10"))
	assert.Equal(t, "/user?id=3", app.Route("user", &Param{Name: "id", Value: 3}))
}

func TestRoutesMerge(t *testing.T) {
	app := New()
	app.Routes(Routes{"a": "/1"})
	app.Routes(Routes{"b": "/2"})
	app.Routes(Routes{"a": "/3"})

	assert.Equal(t, "/3", app.Route("a"))
	assert.Equal(t, "/2", app.Route("b"))
}

func TestRouteNotFoundErrorType(t *testing.T) {
	app := New()
	app.Routes(Routes{"a": "/1"})

	defer func() {
		r := recover()
		assert.IsType(t, &ErrRouteNotFound{}, r)
		if err, ok := r.(error); assert.True(t, ok) {
			assert.Contains(t, err.Error(), "route 'missing' not found")
		}
	}()
	app.Route("missing")
	t.Fatal("expected panic")
}

func TestRouteFuncWithoutApp(t *testing.T) {
	assert.PanicsWithValue(t, ErrAppNotFound, func() {
		Route(context.Background(), "x")
	})
}

func TestContextIsRoute(t *testing.T) {
	app := New()
	app.Routes(Routes{
		"home":     "/",
		"admin":    "/admin",
		"users":    "/admin/users",
		"filtered": "/items?tab=1",
		"docs":     "/docs/", // trailing slash is normalized away
	})

	isRoute := func(path, name string) bool {
		r := httptest.NewRequest("GET", path, nil)
		return NewAppContext(app, httptest.NewRecorder(), r).IsRoute(name)
	}

	t.Run("exact match", func(t *testing.T) {
		assert.True(t, isRoute("/admin", "admin"))
	})

	t.Run("most specific route wins on a sub-path", func(t *testing.T) {
		assert.True(t, isRoute("/admin/users/42", "users"))
		assert.False(t, isRoute("/admin/users/42", "admin")) // deeper route matches
		assert.False(t, isRoute("/admin/users/42", "home"))
	})

	t.Run("section matches its sub-paths when nothing deeper does", func(t *testing.T) {
		assert.True(t, isRoute("/admin/settings", "admin"))
	})

	t.Run("segment aware", func(t *testing.T) {
		assert.False(t, isRoute("/administrators", "admin"))
	})

	t.Run("root only on exact path", func(t *testing.T) {
		assert.True(t, isRoute("/", "home"))
		assert.False(t, isRoute("/admin", "home"))
	})

	t.Run("query is ignored in both route and request", func(t *testing.T) {
		assert.True(t, isRoute("/items?x=1", "filtered"))
	})

	t.Run("trailing slash in route path is ignored", func(t *testing.T) {
		assert.True(t, isRoute("/docs", "docs"))       // matches without the slash
		assert.True(t, isRoute("/docs/intro", "docs")) // and on sub-paths
	})

	t.Run("panics on unknown route", func(t *testing.T) {
		assert.Panics(t, func() { isRoute("/", "missing") })
	})
}

func TestContextIsRouteInTemplate(t *testing.T) {
	app := New()
	app.Routes(Routes{"home": "/", "about": "/about"})
	app.Template().Parse("nav", `{{if .IsRoute "about"}}active{{end}}`)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/about", nil)
	ctx := NewAppContext(app, w, r)

	// passing ctx as the view data lets the template call its methods
	assert.NoError(t, ctx.View("nav", ctx))
	assert.Equal(t, "active", w.Body.String())
}
