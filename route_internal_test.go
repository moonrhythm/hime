package hime

import (
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("Route internal", func() {
	g.Specify("clone nil return nil", func() {
		Expect(cloneRoutes(nil)).To(BeNil())
	})

	g.Context("given routes with data", func() {
		var (
			t Routes
		)

		g.BeforeEach(func() {
			t = Routes{
				"a": "/b",
				"b": "/cd",
			}
		})

		g.When("clone that routes", func() {
			var (
				p Routes
			)

			g.BeforeEach(func() {
				p = cloneRoutes(t)
			})

			g.Specify("cloned not identical to original", func() {
				Expect(p).NotTo(BeIdenticalTo(t))
			})
		})
	})
})
