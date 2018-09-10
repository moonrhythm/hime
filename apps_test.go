package hime

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApps(t *testing.T) {
	t.Run("Config", func(t *testing.T) {
		var gs GracefulShutdown
		gs.Timeout(10 * time.Second)
		gs.Wait(5 * time.Second)

		apps := Merge()
		apps.Config(AppsConfig{
			GracefulShutdown: &gs,
		})

		assert.Equal(t, apps.gs, &gs)
	})

	t.Run("ParseConfig YAML", func(t *testing.T) {
		apps := Merge()
		apps.ParseConfig([]byte(`
gracefulShutdown:
  timeout: 5s
  wait: 10s`))
		assert.NotNil(t, apps.gs)
	})

	t.Run("ParseConfig invalid YAML", func(t *testing.T) {
		apps := Merge()

		assert.Panics(t, func() {
			apps.ParseConfig([]byte(`
	gracefulShutdown:
			 timeout: 5s
		wait: 10s`))
		})
	})

	t.Run("ParseConfig JSON", func(t *testing.T) {
		apps := Merge()
		apps.ParseConfig([]byte(`{"gracefulShutdown": {"timeout":"5s","wait":"10s"}}`))
		assert.NotNil(t, apps.gs)
	})

	t.Run("ParseConfigFile", func(t *testing.T) {
		apps := Merge()

		assert.NotPanics(t, func() {
			apps.ParseConfigFile("testdata/apps.yaml")
		})
	})

	t.Run("ParseConfigFile not exists", func(t *testing.T) {
		apps := Merge()

		assert.Panics(t, func() {
			apps.ParseConfigFile("testdata/not-exists-apps.yaml")
		})
	})

	t.Run("GracefulShutdown", func(t *testing.T) {
		apps := Merge()

		assert.Nil(t, apps.gs)
		apps.GracefulShutdown()
		assert.NotNil(t, apps.gs)
	})

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

		apps.Shutdown(context.Background())
		assert.Equal(t, called, 2)
	})

	t.Run("ListenAndServe graceful shutdown", func(t *testing.T) {
		t.Parallel()

		h := Handler(func(ctx *Context) error {
			return ctx.String("Hello")
		})

		var (
			called     = false
			app1Called = false
			app2Called = false
		)

		app1 := New().Address(":9093").Handler(h)
		app1.GracefulShutdown().Notify(func() { app1Called = true })
		app2 := New().Address(":9094").Handler(h)
		app2.GracefulShutdown().Notify(func() { app2Called = true })
		apps := Merge(app1, app2)
		gs := apps.GracefulShutdown()
		gs.Wait(time.Second)
		gs.Timeout(time.Second)
		gs.Notify(func() { called = true })

		go apps.ListenAndServe()
		time.Sleep(time.Second)

		http.Get("http://localhost:9093")
		http.Get("http://localhost:9094")

		apps.Shutdown(context.Background())
		time.Sleep(time.Second)

		assert.True(t, called)
		assert.True(t, app1Called)
		assert.True(t, app2Called)
	})

	t.Run("ListenAndServe graceful shutdown error duplicate port", func(t *testing.T) {
		t.Parallel()

		app1 := New().Address(":9095")
		app2 := New().Address(":9095")
		apps := Merge(app1, app2)
		apps.GracefulShutdown()

		assert.Error(t, apps.ListenAndServe())
	})
}
