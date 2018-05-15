package hime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	var err error

	err = newErrRouteNotFound("route123")
	assert.IsType(t, &ErrRouteNotFound{}, err)
	assert.Contains(t, err.Error(), "route123")

	err = newErrTemplateDuplicate("temp123")
	assert.IsType(t, &ErrTemplateDuplicate{}, err)
	assert.Contains(t, err.Error(), "temp123")

	err = newErrTemplateNotFound("temp123")
	assert.IsType(t, &ErrTemplateNotFound{}, err)
	assert.Contains(t, err.Error(), "temp123")
}
