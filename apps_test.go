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
}
