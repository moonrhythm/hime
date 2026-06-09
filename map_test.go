package hime

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mapLen(m *sync.Map) (i int) {
	m.Range(func(_, _ any) bool {
		i++
		return true
	})
	return
}

func TestCloneMap(t *testing.T) {
	t.Parallel()

	var src sync.Map
	src.Store("k1", "v1")
	src.Store(2, "v2")

	dst := cloneMap(&src)
	assert.Equal(t, 2, mapLen(&dst))

	v1, ok := dst.Load("k1")
	assert.True(t, ok)
	assert.Equal(t, "v1", v1)
	v2, ok := dst.Load(2)
	assert.True(t, ok)
	assert.Equal(t, "v2", v2)

	// clone is independent of the source
	dst.Store("k3", "v3")
	_, ok = src.Load("k3")
	assert.False(t, ok)
}
