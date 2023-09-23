package hime

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApps(t *testing.T) {
	t.Run("ListenAndServe", func(t *testing.T) {
		t.Parallel()

		called := 0
		h := Handler(func(ctx *Context) error {
			called++
			return ctx.String("Hello")
		})

		app1 := New().Address(":9091").Handler(h)
		app2 := New().Address(":9092").Handler(h)
		apps := Merge(app1, app2)

		go apps.ListenAndServe()
		time.Sleep(time.Second)

		http.Get("http://localhost:9091")
		http.Get("http://localhost:9092")

		apps.Shutdown()
		assert.Equal(t, called, 2)
	})

	t.Run("ListenAndServe error duplicate port", func(t *testing.T) {
		t.Parallel()

		app1 := New().Address(":9095")
		app2 := New().Address(":9095")
		apps := Merge(app1, app2)

		assert.Error(t, apps.ListenAndServe())
	})
}
