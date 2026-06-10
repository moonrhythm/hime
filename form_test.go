package hime_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moonrhythm/hime"
)

func newFormContext(body string) *hime.Context {
	r := httptest.NewRequest(http.MethodPost, "/?q=z", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return hime.NewAppContext(hime.New(), httptest.NewRecorder(), r)
}

func TestFormStateValues(t *testing.T) {
	t.Parallel()

	fs := newFormContext("email=a@b.com&tag=x&tag=y").FormState()

	assert.Equal(t, "a@b.com", fs.Value("email"))
	assert.Equal(t, []string{"x", "y"}, fs.Values("tag"))
	assert.Equal(t, "z", fs.Value("q")) // query values are included
	assert.Equal(t, "", fs.Value("missing"))
	assert.Empty(t, fs.Values("missing"))
}

func TestFormStateErrors(t *testing.T) {
	t.Parallel()

	fs := newFormContext("email=bad").FormState()

	assert.False(t, fs.HasErrors())
	assert.False(t, fs.HasError("email"))
	assert.Equal(t, "", fs.Error("email"))
	assert.Empty(t, fs.Errors("email"))

	fs.AddError("email", "is taken")
	fs.AddError("email", "is invalid")

	assert.True(t, fs.HasErrors())
	assert.True(t, fs.HasError("email"))
	assert.False(t, fs.HasError("name"))
	assert.Equal(t, "is taken", fs.Error("email"))
	assert.Equal(t, []string{"is taken", "is invalid"}, fs.Errors("email"))
}

func TestFormStateSetValue(t *testing.T) {
	t.Parallel()

	fs := newFormContext("email=a@b.com").FormState()
	fs.SetValue("email", "c@d.com")
	fs.SetValue("name", "hime")

	assert.Equal(t, "c@d.com", fs.Value("email"))
	assert.Equal(t, "hime", fs.Value("name"))
}

func TestFormStateDoesNotMutateRequest(t *testing.T) {
	t.Parallel()

	ctx := newFormContext("email=a@b.com")
	fs := ctx.FormState()
	fs.SetValue("email", "changed")

	// the request's own form value is untouched by the copy
	assert.Equal(t, "a@b.com", ctx.FormValue("email"))
}
