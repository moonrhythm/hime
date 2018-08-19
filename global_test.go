package hime_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/acoshift/hime"
)

var _ = Describe("Global", func() {
	Describe("given new app with globals data", func() {
		var (
			app *hime.App
		)

		BeforeEach(func() {
			app = hime.New()

			app.Globals(hime.Globals{
				"key1": "value1",
				"key2": "value2",
			})
		})

		It("should be able to retrieve globals data from app", func() {
			Expect(app.Global("key1")).To(Equal("value1"))
			Expect(app.Global("key2")).To(Equal("value2"))
			Expect(app.Global("key3")).To(BeNil())
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

			It("should be able to retrieve globals data from handler", func() {
				app.Handler(hime.Handler(func(ctx *hime.Context) error {
					Expect(ctx.Global("key1")).To(Equal("value1"))
					Expect(ctx.Global("key2")).To(Equal("value2"))
					Expect(ctx.Global("key3")).To(BeNil())
					return nil
				})).ServeHTTP(w, r)
			})
		})
	})
})
