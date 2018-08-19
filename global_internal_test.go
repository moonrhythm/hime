package hime

import (
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("Global internal", func() {
	g.Specify("clone nil return nil", func() {
		Expect(cloneGlobals(nil)).To(BeNil())
	})

	g.Context("given globals with data", func() {
		var (
			t Globals
		)

		g.BeforeEach(func() {
			t = Globals{
				"a": 1,
				"b": 2,
			}
		})

		g.When("clone that globals", func() {
			var (
				p Globals
			)

			g.BeforeEach(func() {
				p = cloneGlobals(t)
			})

			g.Specify("cloned not identical to original", func() {
				Expect(p).NotTo(BeIdenticalTo(t))
			})
		})
	})
})
