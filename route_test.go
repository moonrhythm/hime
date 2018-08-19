package hime_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/acoshift/hime"
)

var _ = Describe("Route", func() {
	Describe("given new app without routes data", func() {
		var (
			app *hime.App
		)

		BeforeEach(func() {
			app = hime.New()
		})

		It("should panic when retrieve any route", func() {
			Expect(func() { app.Route("r1") }).Should(Panic())
		})
	})

	Describe("given new app with routes", func() {
		var (
			app *hime.App
		)

		BeforeEach(func() {
			app = hime.New()

			app.Routes(hime.Routes{
				"a": "/b",
				"b": "/cd",
			})
		})

		It("should be able to retrieve route from app", func() {
			Expect(app.Route("a")).To(Equal("/b"))
			Expect(app.Route("b")).To(Equal("/cd"))
		})

		It("should panic when retrieve not exist route from app", func() {
			Expect(func() { app.Route("c") }).Should(Panic())
		})

		When("calling app with a handler", func() {
			var (
				w *httptest.ResponseRecorder
				r *http.Request
			)

			BeforeEach(func() {
				w = httptest.NewRecorder()
				r = httptest.NewRequest("GET", "/", nil)
			})

			It("should be able to retrieve route from handler", func() {
				app.Handler(hime.Handler(func(ctx *hime.Context) error {
					Expect(ctx.Route("a")).To(Equal("/b"))
					Expect(ctx.Route("b")).To(Equal("/cd"))
					return nil
				})).ServeHTTP(w, r)
			})

			It("should panic when retrieve not exist route from handler", func() {
				app.Handler(hime.Handler(func(ctx *hime.Context) error {
					Expect(func() { app.Route("c") }).Should(Panic())
					return nil
				})).ServeHTTP(w, r)
			})
		})
	})
})
