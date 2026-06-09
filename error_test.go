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

func TestErrorMessages(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "hime: route 'r' not found", newErrRouteNotFound("r").Error())
	assert.Equal(t, "hime: template 't' not found", newErrTemplateNotFound("t").Error())
	assert.Equal(t, "hime: template 't' already exists", newErrTemplateDuplicate("t").Error())
	assert.Equal(t, "hime: component 'c' not found", newErrComponentNotFound("c").Error())
	assert.Equal(t, "hime: component 'c' already exists", newErrComponentDuplicate("c").Error())
	assert.Equal(t, "hime: app not found", ErrAppNotFound.Error())
}

func TestPanicf(t *testing.T) {
	t.Parallel()

	// panicf prefixes "hime: " and formats the message.
	assert.PanicsWithValue(t, "hime: bad value 42", func() { panicf("bad value %d", 42) })
}
