package hime_test

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/acoshift/hime"
)

var _ = Describe("Context", func() {
	var (
		w *httptest.ResponseRecorder
		r *http.Request
	)

	BeforeEach(func() {
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/", nil)
	})

	It("should panic when create context without App", func() {
		Expect(func() { hime.NewContext(w, r) }).Should(Panic())
	})

	Describe("context data", func() {
		When("create context with an App", func() {
			var (
				app *hime.App
				ctx *hime.Context
			)

			BeforeEach(func() {
				app = hime.New()
				ctx = hime.NewAppContext(app, w, r)
			})

			Specify("an Request() is given request", func() {
				Expect(ctx.Request()).To(Equal(r))
			})

			Specify("an ResponseWriter() is given response writer", func() {
				Expect(ctx.ResponseWriter()).To(Equal(w))
			})

			Specify("ctx.Param is short hand for hime.Param", func() {
				Expect(ctx.Param("id", 11)).To(Equal(&hime.Param{Name: "id", Value: 11}))
			})

			When("inject value to context", func() {
				BeforeEach(func() {
					ctx.WithValue("data", "text")
				})

				It("should be able to retrieve that value", func() {
					Expect(ctx.Value("data")).To(Equal("text"))
				})
			})

			When("override request", func() {
				var (
					nr *http.Request
				)

				BeforeEach(func() {
					nr = httptest.NewRequest(http.MethodPost, "/test", nil)
					ctx.WithRequest(nr)
				})

				Specify("an Request() is new request", func() {
					Expect(ctx.Request()).To(BeIdenticalTo(nr))
					Expect(ctx.Request()).NotTo(BeIdenticalTo(r))
				})
			})

			When("override response writer", func() {
				var (
					nw http.ResponseWriter
				)

				BeforeEach(func() {
					nw = httptest.NewRecorder()
					ctx.WithResponseWriter(nw)
				})

				Specify("an ResponseWriter() is new request", func() {
					Expect(ctx.ResponseWriter()).To(BeIdenticalTo(nw))
					Expect(ctx.ResponseWriter()).NotTo(BeIdenticalTo(w))
				})
			})

			When("given deadline context", func() {
				var (
					nctx   context.Context
					cancel context.CancelFunc
				)

				BeforeEach(func() {
					nctx, cancel = context.WithTimeout(ctx, 5*time.Second)
					ctx.WithContext(nctx)
				})

				AfterEach(func() {
					cancel()
				})

				Specify("deadline is a given context deadline", func() {
					t, ok := ctx.Deadline()
					nt, nok := nctx.Deadline()
					Expect(t).To(Equal(nt))
					Expect(ok).To(Equal(nok))
				})

				Specify("done is a given context done", func() {
					Expect(ctx.Done()).To(Equal(nctx.Done()))
				})

				When("cancel context", func() {
					BeforeEach(func() {
						cancel()
					})

					Specify("error is a given context error", func() {
						Expect(ctx.Err()).ToNot(BeNil())
						Expect(ctx.Err()).To(Equal(nctx.Err()))
					})
				})
			})
		})
	})

	Describe("context response", func() {
		When("create context with an App", func() {
			var (
				app *hime.App
				ctx *hime.Context
			)

			BeforeEach(func() {
				app = hime.New()
				ctx = hime.NewAppContext(app, w, r)
			})

			Describe("testing Status", func() {
				When("set status code to 200", func() {
					BeforeEach(func() {
						ctx.Status(200).StatusText()
					})

					It("should response with 200 status code", func() {
						Expect(w.Code).To(Equal(200))
					})
				})

				When("set status code to 400", func() {
					BeforeEach(func() {
						ctx.Status(400).StatusText()
					})

					It("should response with 400 status code", func() {
						Expect(w.Code).To(Equal(400))
					})
				})

				When("set status code to 500", func() {
					BeforeEach(func() {
						ctx.Status(500).StatusText()
					})

					It("should response with 500 status code", func() {
						Expect(w.Code).To(Equal(500))
					})
				})
			})

			Describe("testing StatusText", func() {
				When("response with status text", func() {
					BeforeEach(func() {
						ctx.Status(http.StatusTeapot).StatusText()
					})

					Specify("responsed body to be status text", func() {
						Expect(w.Body.String()).To(Equal(http.StatusText(http.StatusTeapot)))
					})
				})
			})

			Describe("testing NotFound", func() {
				When("calling NotFound", func() {
					BeforeEach(func() {
						ctx.NotFound()
					})

					Specify("responsed status code to be 404", func() {
						Expect(w.Code).To(Equal(404))
					})

					Specify("responsed body to be not found", func() {
						Expect(w.Body.String()).To(Equal("404 page not found\n"))
					})
				})
			})

			Describe("testing NoContent", func() {
				When("calling NoContent", func() {
					BeforeEach(func() {
						ctx.NoContent()
					})

					Specify("responsed status code to be 204", func() {
						Expect(w.Code).To(Equal(204))
					})

					Specify("responsed body to be empty", func() {
						Expect(w.Body.String()).To(BeEmpty())
					})
				})
			})

			Describe("testing Bytes", func() {
				When("calling Bytes with a data", func() {
					BeforeEach(func() {
						ctx.Bytes([]byte("hello hime"))
					})

					Specify("responsed status code to be 200", func() {
						Expect(w.Code).To(Equal(200))
					})

					Specify("responsed content type to be application/octet-stream", func() {
						Expect(w.Header().Get("Content-Type")).To(Equal("application/octet-stream"))
					})

					Specify("responsed body to be the response data", func() {
						Expect(w.Body.String()).To(Equal("hello hime"))
					})
				})
			})

			Describe("testing File", func() {
				When("calling File with a text file", func() {
					BeforeEach(func() {
						ctx.File("testdata/file.txt")
					})

					Specify("responsed status code to be 200", func() {
						Expect(w.Code).To(Equal(200))
					})

					Specify("responsed content type to be text/plain", func() {
						Expect(w.Header().Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))
					})

					Specify("responsed body to be the file content", func() {
						Expect(w.Body.String()).To(Equal("file content"))
					})
				})
			})

			Describe("testing JSON", func() {
				When("calling JSON with a data", func() {
					BeforeEach(func() {
						ctx.JSON(map[string]interface{}{"abc": "afg", "bbb": 123})
					})

					Specify("responsed status code to be 200", func() {
						Expect(w.Code).To(Equal(200))
					})

					Specify("responsed content type to be application/json", func() {
						Expect(w.Header().Get("Content-Type")).To(Equal("application/json; charset=utf-8"))
					})

					Specify("responsed body to be the json data", func() {
						Expect(w.Body.String()).To(MatchJSON(`{"abc":"afg","bbb":123}`))
					})
				})
			})

			Describe("testing HTML", func() {
				When("calling HTML with a data", func() {
					BeforeEach(func() {
						ctx.HTML([]byte(`<h1>Hello</h1>`))
					})

					Specify("responsed status code to be 200", func() {
						Expect(w.Code).To(Equal(200))
					})

					Specify("responsed content type to be text/html", func() {
						Expect(w.Header().Get("Content-Type")).To(Equal("text/html; charset=utf-8"))
					})

					Specify("responsed body to be the html data", func() {
						Expect(w.Body.String()).To(Equal(`<h1>Hello</h1>`))
					})
				})
			})

			Describe("testing String", func() {
				When("calling String with a data", func() {
					BeforeEach(func() {
						ctx.String("hello, hime")
					})

					Specify("responsed status code to be 200", func() {
						Expect(w.Code).To(Equal(200))
					})

					Specify("responsed content type to be text/plain", func() {
						Expect(w.Header().Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))
					})

					Specify("responsed body to be the text data", func() {
						Expect(w.Body.String()).To(Equal("hello, hime"))
					})
				})
			})

			Describe("testing Error", func() {
				When("calling Error with an error", func() {
					BeforeEach(func() {
						ctx.Error("some error")
					})

					Specify("responsed status code to be 500", func() {
						Expect(w.Code).To(Equal(500))
					})

					Specify("responsed content type to be text/plain", func() {
						Expect(w.Header().Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))
					})

					Specify("responsed body to be the error", func() {
						Expect(w.Body.String()).To(Equal("some error\n"))
					})
				})

				When("calling Error with 404 status code", func() {
					BeforeEach(func() {
						ctx.Status(http.StatusNotFound).Error("some error")
					})

					Specify("responsed status code to be 404", func() {
						Expect(w.Code).To(Equal(404))
					})

					Specify("responsed content type to be text/plain", func() {
						Expect(w.Header().Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))
					})

					Specify("responsed body to be the error", func() {
						Expect(w.Body.String()).To(Equal("some error\n"))
					})
				})
			})

			Describe("testing Redirect", func() {
				When("calling Redirect with an external url", func() {
					BeforeEach(func() {
						ctx.Redirect("https://google.com")
					})

					Specify("responsed status code to be 302", func() {
						Expect(w.Code).To(Equal(302))
					})

					Specify("responsed location should be the redirect location", func() {
						Expect(w.Header().Get("Location")).To(Equal("https://google.com"))
					})
				})

				When("calling Redirect with an internal url path", func() {
					BeforeEach(func() {
						ctx.Redirect("/signin")
					})

					Specify("responsed status code to be 302", func() {
						Expect(w.Code).To(Equal(302))
					})

					Specify("responsed location should be the redirect location", func() {
						Expect(w.Header().Get("Location")).To(Equal("/signin"))
					})
				})

				When("calling Redirect with 301 status code", func() {
					BeforeEach(func() {
						ctx.Status(301).Redirect("/signin")
					})

					Specify("responsed status code to be 301", func() {
						Expect(w.Code).To(Equal(301))
					})

					Specify("responsed location should be the redirect location", func() {
						Expect(w.Header().Get("Location")).To(Equal("/signin"))
					})
				})

				When("calling Redirect with request POST method", func() {
					BeforeEach(func() {
						ctx.Request().Method = "POST"
						ctx.Redirect("/signin")
					})

					Specify("responsed status code to be 303", func() {
						Expect(w.Code).To(Equal(303))
					})

					Specify("responsed location should be the redirect location", func() {
						Expect(w.Header().Get("Location")).To(Equal("/signin"))
					})
				})
			})

			Describe("testing RedirectTo", func() {
				Context("given routes to the app", func() {
					BeforeEach(func() {
						app.Routes(hime.Routes{
							"route1": "/route/1",
						})
					})

					When("calling RedirectTo with valid route", func() {
						BeforeEach(func() {
							ctx.RedirectTo("route1")
						})

						Specify("responsed status code to be 302", func() {
							Expect(w.Code).To(Equal(302))
						})

						Specify("responsed location should be the redirect location", func() {
							Expect(w.Header().Get("Location")).To(Equal("/route/1"))
						})
					})

					When("calling RedirectTo with valid route and param", func() {
						BeforeEach(func() {
							ctx.RedirectTo("route1", ctx.Param("id", 3))
						})

						Specify("responsed status code to be 302", func() {
							Expect(w.Code).To(Equal(302))
						})

						Specify("responsed location should be the redirect location", func() {
							Expect(w.Header().Get("Location")).To(Equal("/route/1?id=3"))
						})
					})

					When("calling RedirectTo with additional path", func() {
						BeforeEach(func() {
							ctx.RedirectTo("route1", "create")
						})

						Specify("responsed status code to be 302", func() {
							Expect(w.Code).To(Equal(302))
						})

						Specify("responsed location should be the redirect location", func() {
							Expect(w.Header().Get("Location")).To(Equal("/route/1/create"))
						})
					})

					When("calling RedirectTo with additional path and param", func() {
						BeforeEach(func() {
							ctx.RedirectTo("route1", "create", ctx.Param("id", 3))
						})

						Specify("responsed status code to be 302", func() {
							Expect(w.Code).To(Equal(302))
						})

						Specify("responsed location should be the redirect location", func() {
							Expect(w.Header().Get("Location")).To(Equal("/route/1/create?id=3"))
						})
					})

					When("calling RedirectTo with 301 status code", func() {
						BeforeEach(func() {
							ctx.Status(301).RedirectTo("route1")
						})

						Specify("responsed status code to be 301", func() {
							Expect(w.Code).To(Equal(301))
						})

						Specify("responsed location should be the redirect location", func() {
							Expect(w.Header().Get("Location")).To(Equal("/route/1"))
						})
					})

					It("should panic when calling RedirectTo with invalid route", func() {
						Expect(func() { ctx.RedirectTo("invalid") }).Should(Panic())
					})
				})
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

			Describe("testing Handle", func() {
				It("should invoke given handler when calling Handle", func() {
					called := false
					ctx.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						called = true
					}))

					Expect(called).To(BeTrue())
				})
			})
		})
	})
})
