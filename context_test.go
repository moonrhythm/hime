package hime_test

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/moonrhythm/hime"
)

func TestContext(t *testing.T) {
	t.Parallel()

	t.Run("panic when create context without app", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		assert.Panics(t, func() { hime.NewContext(w, r) })
	})

	t.Run("basic data", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.Equal(t, ctx.Request, r, "ctx.Request must be given request")
		assert.Equal(t, ctx.ResponseWriter(), w, "ctx.ResponseWriter() must return given response writer")
		assert.Equal(t, ctx.Param("id", 11), &hime.Param{Name: "id", Value: 11}, "ctx.Param must returns a Param")
	})

	t.Run("Value", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx.WithValue("data", "text")
		assert.Equal(t, ctx.Value("data"), "text")
	})

	t.Run("WithResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		nw := httptest.NewRecorder()
		ctx.WithResponseWriter(nw)
		assert.Equal(t, ctx.ResponseWriter(), nw)
	})

	t.Run("Deadline", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		nctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		ctx.WithContext(nctx)

		dt, ok := ctx.Deadline()
		ndt, nok := nctx.Deadline()
		assert.Equal(t, dt, ndt)
		assert.Equal(t, ok, nok)
		assert.Equal(t, ctx.Done(), nctx.Done())

		cancel()
		assert.Error(t, ctx.Err())
		assert.Equal(t, ctx.Err(), nctx.Err())
	})

	t.Run("Handle", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		called := false
		assert.NoError(t, ctx.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})))

		assert.True(t, called)
	})

	t.Run("AddHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx.AddHeader("Vary", "b")
		assert.Equal(t, w.Header().Get("Vary"), "b")
	})

	t.Run("AddHeaderIfNotExists", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx.AddHeaderIfNotExists("Vary", "b")
		ctx.AddHeaderIfNotExists("Vary", "c")
		assert.Len(t, w.Header()["Vary"], 1)
		assert.Equal(t, w.Header().Get("Vary"), "b")
	})

	t.Run("SetHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx.SetHeader("Vary", "b")
		ctx.SetHeader("Vary", "c")
		assert.Len(t, w.Header()["Vary"], 1)
		assert.Equal(t, w.Header().Get("Vary"), "c")
	})

	t.Run("DelHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		ctx.SetHeader("Vary", "b")
		ctx.DelHeader("Vary")
		assert.Empty(t, w.Header().Get("Vary"))
	})

	t.Run("Status", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(401).StatusText())
		assert.Equal(t, w.Code, 401)
	})

	t.Run("StatusText", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(http.StatusTeapot).StatusText())
		assert.Equal(t, w.Code, http.StatusTeapot)
		assert.Equal(t, w.Body.String(), http.StatusText(http.StatusTeapot))
	})

	t.Run("NoContent", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.NoContent())
		assert.Equal(t, w.Code, http.StatusNoContent)
		assert.Empty(t, w.Body.String())
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.NotFound())
		assert.Equal(t, w.Code, http.StatusNotFound)
		assert.Equal(t, w.Body.String(), "404 page not found\n")
	})

	t.Run("Bytes", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Bytes([]byte("hello hime")))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/octet-stream")
		assert.Equal(t, w.Body.String(), "hello hime")
	})

	t.Run("File", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.File("testdata/file.txt"))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
		assert.Equal(t, w.Body.String(), "file content")
	})

	t.Run("JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.JSON(map[string]interface{}{"abc": "afg", "bbb": 123}))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json; charset=utf-8")
		assert.JSONEq(t, w.Body.String(), `{"abc":"afg","bbb":123}`)
	})

	t.Run("HTML", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.HTML([]byte(`<h1>Hello</h1>`)))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
		assert.Equal(t, w.Body.String(), `<h1>Hello</h1>`)
	})

	t.Run("String", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.String("hello, hime"))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
		assert.Equal(t, w.Body.String(), "hello, hime")
	})

	t.Run("Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Error("some error"))
		assert.Equal(t, w.Code, http.StatusInternalServerError)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
		assert.Equal(t, w.Body.String(), "some error\n")
	})

	t.Run("Error with status", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(http.StatusNotFound).Error("some error"))
		assert.Equal(t, w.Code, http.StatusNotFound)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
		assert.Equal(t, w.Body.String(), "some error\n")
	})

	t.Run("Redirect to external url", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Redirect("https://google.com"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "https://google.com")
	})

	t.Run("Redirect to internal url path", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Redirect("/signin"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/signin")
	})

	t.Run("Redirect with status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(http.StatusMovedPermanently).Redirect("/signin"))
		assert.Equal(t, w.Code, http.StatusMovedPermanently)
		assert.Equal(t, w.Header().Get("Location"), "/signin")
	})

	t.Run("Redirect with PRG", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Redirect("/signin"))
		assert.Equal(t, w.Code, http.StatusSeeOther)
		assert.Equal(t, w.Header().Get("Location"), "/signin")
	})

	t.Run("Redirect with PRG and status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		app := hime.New()
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(http.StatusPermanentRedirect).Redirect("/signin"))
		assert.Equal(t, w.Code, http.StatusPermanentRedirect)
		assert.Equal(t, w.Header().Get("Location"), "/signin")
	})

	t.Run("RedirectTo to valid route", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectTo("route1"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/route/1")
	})

	t.Run("RedirectTo to valid route with param", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectTo("route1", ctx.Param("id", 3)))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/route/1?id=3")
	})

	t.Run("RedirectTo to valid route with additional path", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectTo("route1", "create"))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/route/1/create")
	})

	t.Run("RedirectTo to valid route with additional path and param", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.RedirectTo("route1", "create", ctx.Param("id", 3)))
		assert.Equal(t, w.Code, http.StatusFound)
		assert.Equal(t, w.Header().Get("Location"), "/route/1/create?id=3")
	})

	t.Run("RedirectTo to valid route with status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.NoError(t, ctx.Status(301).RedirectTo("route1"))
		assert.Equal(t, w.Code, http.StatusMovedPermanently)
		assert.Equal(t, w.Header().Get("Location"), "/route/1")
	})

	t.Run("RedirectTo to invalid route", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		app := hime.New()
		app.Routes(hime.Routes{
			"route1": "/route/1",
		})
		ctx := hime.NewAppContext(app, w, r)

		assert.Panics(t, func() { ctx.RedirectTo("invalid") })
	})
}

var _ = Describe("Context", func() {
	var (
		w *httptest.ResponseRecorder
		r *http.Request
	)

	BeforeEach(func() {
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/", nil)
	})

	Describe("context response", func() {
		var (
			app *hime.App
			ctx *hime.Context
		)

		BeforeEach(func() {
			app = hime.New()
			ctx = hime.NewAppContext(app, w, r)
		})

		Describe("testing RedirectToGet", func() {
			When("calling RedirectToGet", func() {
				BeforeEach(func() {
					ctx.RedirectToGet()
				})

				Specify("responsed status code to be 303", func() {
					Expect(w.Code).To(Equal(303))
				})

				Specify("responsed location should be the request uri", func() {
					Expect(w.Header().Get("Location")).To(Equal(r.RequestURI))
				})
			})
		})

		Describe("testing RedirectBack", func() {
			When("calling RedirectBack with empty fallback", func() {
				BeforeEach(func() {
					ctx.RedirectBack("")
				})

				Specify("responsed status code to be 302", func() {
					Expect(w.Code).To(Equal(302))
				})

				Specify("responsed location should be the request uri", func() {
					Expect(w.Header().Get("Location")).To(Equal(r.RequestURI))
				})
			})

			When("calling RedirectBack with a fallback", func() {
				BeforeEach(func() {
					ctx.RedirectBack("/path2")
				})

				Specify("responsed status code to be 302", func() {
					Expect(w.Code).To(Equal(302))
				})

				Specify("responsed location should be the fallback url", func() {
					Expect(w.Header().Get("Location")).To(Equal("/path2"))
				})
			})

			Context("given referer to request", func() {
				BeforeEach(func() {
					r.Header.Set("Referer", "http://localhost/path1")
				})

				When("calling RedirectBack with empty fallback", func() {
					BeforeEach(func() {
						ctx.RedirectBack("")
					})

					Specify("responsed status code to be 302", func() {
						Expect(w.Code).To(Equal(302))
					})

					Specify("responsed location should be the referer url", func() {
						Expect(w.Header().Get("Location")).To(Equal(r.Referer()))
					})
				})

				When("calling RedirectBack with a fallback", func() {
					BeforeEach(func() {
						ctx.RedirectBack("/path2")
					})

					Specify("responsed status code to be 302", func() {
						Expect(w.Code).To(Equal(302))
					})

					Specify("responsed location should still be the referer url", func() {
						Expect(w.Header().Get("Location")).To(Equal(r.Referer()))
					})
				})
			})
		})

		Describe("testing SafeRedirectBack", func() {
			When("calling SafeRedirectBack with empty fallback", func() {
				BeforeEach(func() {
					ctx.SafeRedirectBack("")
				})

				Specify("responsed status code to be 302", func() {
					Expect(w.Code).To(Equal(302))
				})

				Specify("responsed location should be the request uri", func() {
					Expect(w.Header().Get("Location")).To(Equal(r.RequestURI))
				})
			})

			When("calling SafeRedirectBack with a fallback", func() {
				BeforeEach(func() {
					ctx.SafeRedirectBack("/path2")
				})

				Specify("responsed status code to be 302", func() {
					Expect(w.Code).To(Equal(302))
				})

				Specify("responsed location should be the fallback url", func() {
					Expect(w.Header().Get("Location")).To(Equal("/path2"))
				})
			})

			When("calling SafeRedirectBack with a dangerous fallback", func() {
				BeforeEach(func() {
					ctx.SafeRedirectBack("https://google.com/path2")
				})

				Specify("responsed status code to be 302", func() {
					Expect(w.Code).To(Equal(302))
				})

				Specify("responsed location should be the safe fallback url", func() {
					Expect(w.Header().Get("Location")).To(Equal("/path2"))
				})
			})

			Context("given referer to request", func() {
				BeforeEach(func() {
					r.Header.Set("Referer", "http://localhost/path1")
				})

				When("calling SafeRedirectBack with empty fallback", func() {
					BeforeEach(func() {
						ctx.SafeRedirectBack("")
					})

					Specify("responsed status code to be 302", func() {
						Expect(w.Code).To(Equal(302))
					})

					Specify("responsed location should be the safe referer url", func() {
						Expect(w.Header().Get("Location")).To(Equal(hime.SafeRedirectPath(r.Referer())))
					})
				})

				When("calling SafeRedirectBack with a fallback", func() {
					BeforeEach(func() {
						ctx.SafeRedirectBack("/path2")
					})

					Specify("responsed status code to be 302", func() {
						Expect(w.Code).To(Equal(302))
					})

					Specify("responsed location should still be the safe referer url", func() {
						Expect(w.Header().Get("Location")).To(Equal(hime.SafeRedirectPath(r.Referer())))
					})
				})
			})
		})

		Describe("testing RedirectBackToGet", func() {
			When("calling RedirectBackToGet", func() {
				BeforeEach(func() {
					ctx.RedirectBackToGet()
				})

				Specify("responsed status code to be 303", func() {
					Expect(w.Code).To(Equal(303))
				})

				Specify("responsed location should be the request uri", func() {
					Expect(w.Header().Get("Location")).To(Equal(r.RequestURI))
				})
			})

			Context("given referer to request", func() {
				BeforeEach(func() {
					r.Header.Set("Referer", "http://localhost/path1")
				})

				When("calling RedirectBackToGet", func() {
					BeforeEach(func() {
						ctx.RedirectBackToGet()
					})

					Specify("responsed status code to be 303", func() {
						Expect(w.Code).To(Equal(303))
					})

					Specify("responsed location should be the referer", func() {
						Expect(w.Header().Get("Location")).To(Equal(r.Referer()))
					})
				})
			})
		})

		Describe("testing SafeRedirect", func() {
			When("calling SafeRedirect", func() {
				BeforeEach(func() {
					ctx.SafeRedirect("https://google.com")
				})

				Specify("responsed status code to be 302", func() {
					Expect(w.Code).To(Equal(302))
				})

				Specify("responsed location should be safe path", func() {
					Expect(w.Header().Get("Location")).To(Equal("/"))
				})
			})
		})

		Describe("testing View", func() {
			It("should panic when calling View with not exist template", func() {
				Expect(func() { ctx.View("invalid", nil) }).Should(Panic())
			})

			Context("given a view to the app", func() {
				BeforeEach(func() {
					app.Template().Dir("testdata").Root("root").ParseFiles("index", "hello.tmpl")
				})

				When("calling View with valid template", func() {
					BeforeEach(func() {
						ctx.View("index", nil)
					})

					Specify("responsed status code to be 200", func() {
						Expect(w.Code).To(Equal(200))
					})

					Specify("responsed content type to be text/html", func() {
						Expect(w.Header().Get("Content-Type")).To(Equal("text/html; charset=utf-8"))
					})

					Specify("responsed body to be the template data", func() {
						Expect(w.Body.String()).To(Equal("hello"))
					})
				})

				When("calling View with valid template and 500 status code", func() {
					BeforeEach(func() {
						ctx.Status(500).View("index", nil)
					})

					Specify("responsed status code to be 500", func() {
						Expect(w.Code).To(Equal(500))
					})

					Specify("responsed content type to be text/html", func() {
						Expect(w.Header().Get("Content-Type")).To(Equal("text/html; charset=utf-8"))
					})

					Specify("responsed body to be the template data", func() {
						Expect(w.Body.String()).To(Equal("hello"))
					})
				})
			})

			Context("given template funcs to the app", func() {
				BeforeEach(func() {
					app.TemplateFuncs(template.FuncMap{
						"fn":    func(s string) string { return s },
						"panic": func() string { panic("panic") },
					})
				})

				Context("given a template that invoke wrong template func argument", func() {
					BeforeEach(func() {
						app.Template().Dir("testdata").Root("root").ParseFiles("index", "call_fn.tmpl")
					})

					Specify("an error to be return calling View", func() {
						Expect(ctx.View("index", nil)).ToNot(BeNil())
					})
				})

				Context("given a template that invoke panic template func", func() {
					BeforeEach(func() {
						app.Template().Dir("testdata").Root("root").ParseFiles("index", "panic.tmpl")
					})

					It("should panic when calling View", func() {
						Expect(func() { ctx.View("index", nil) }).Should(Panic())
					})
				})
			})
		})
	})
})
