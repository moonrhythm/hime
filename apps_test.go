package hime

import (
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

		assert.Equal(t, &gs, apps.gs)
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

	t.Run("GracefulShutdown", func(t *testing.T) {
		apps := Merge()

		assert.Nil(t, apps.gs)
		apps.GracefulShutdown()
		assert.NotNil(t, apps.gs)
	})
}
