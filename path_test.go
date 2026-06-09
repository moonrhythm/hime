package hime

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input  string
		Output string
	}{
		{"", "/"},
		{"/", "/"},
		{"/p", "/p"},
		{"/p/123", "/p/123"},
		{"https://google.com", "https://google.com/"},
		{"https://google.com/test", "https://google.com/test"},
		{"https://google.com/test/", "https://google.com/test/"},
		{"https://google.com/test?p=1", "https://google.com/test?p=1"},
		{"http://google.com/test?p=1", "http://google.com/test?p=1"},
		{"//a?p=1", "//a/?p=1"},
		{"app:///a?p=1", "app:///a?p=1"},
	}

	for _, c := range cases {
		assert.Equal(t, c.Output, buildPath(c.Input))
	}
}

func TestBuildPathParams(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Base   string
		Params []any
		Output string
	}{
		{"", []any{""}, "/"},
		{"/", []any{"/"}, "/"},
		{"/a", []any{}, "/a"},
		{"/a", []any{"/b"}, "/a/b"},
		{"/a?x=1", []any{"/b"}, "/a/b?x=1"},
		{"/a/", []any{"/b/", "/c/"}, "/a/b/c"},
		{"/a", []any{url.Values{"id": []string{"10"}}}, "/a?id=10"},
		{"/a", []any{"/b", url.Values{"id": []string{"10"}}}, "/a/b?id=10"},
		{"/a", []any{"/b/", url.Values{"id": []string{"10"}}}, "/a/b?id=10"},
		{"/a", []any{"/b/", map[string]string{"id": "10"}}, "/a/b?id=10"},
		{"/a", []any{"/b/", map[string]any{"id": 10}}, "/a/b?id=10"},
		{"/a", []any{"/b", &Param{Name: "id", Value: 3456}}, "/a/b?id=3456"},
		{"/a?x=1", []any{"/b", &Param{Name: "id", Value: 3456}}, "/a/b?id=3456&x=1"},
	}

	for _, c := range cases {
		assert.Equal(t, c.Output, buildPath(c.Base, c.Params...))
	}
}

func TestSafeRedirectPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input  string
		Output string
	}{
		{"", "/"},
		{"/", "/"},
		{"/p", "/p"},
		{"/p/123", "/p/123"},
		{"https://google.com", "/"},
		{"https://google.com/test", "/test"},
		{"https://google.com/test?p=1", "/test?p=1"},
		{"http://google.com/test?p=1", "/test?p=1"},
		{"//a?p=1", "/a?p=1"},
		{"app:///a?p=1", "/a?p=1"},
		{"/p/123?id=3", "/p/123?id=3"},
		{"/p/123/?id=3", "/p/123/?id=3"},
	}

	for _, c := range cases {
		assert.Equal(t, c.Output, SafeRedirectPath(c.Input))
	}
}

func TestBuildPathInvalidURL(t *testing.T) {
	t.Parallel()

	// invalid percent-encoding makes url.Parse fail, which buildPath turns
	// into a panic.
	assert.Panics(t, func() { buildPath("%zz") })
}

func TestSafeRedirectPathInvalid(t *testing.T) {
	t.Parallel()

	// url.ParseRequestURI rejects the input, so SafeRedirectPath defaults to "/".
	assert.Equal(t, "/", SafeRedirectPath("%zz"))
}

func TestSafeRedirectPathClean(t *testing.T) {
	t.Parallel()

	// SafeRedirectPath runs path.Clean, normalizing traversal sequences.
	cases := []struct {
		Input  string
		Output string
	}{
		{"/a/../../etc/passwd", "/etc/passwd"},
		{"/a/./b", "/a/b"},
		{"/foo/..", "/"},
	}
	for _, c := range cases {
		assert.Equal(t, c.Output, SafeRedirectPath(c.Input), "input: %s", c.Input)
	}
}

func TestBuildPathParamMultiValue(t *testing.T) {
	t.Parallel()

	// Same query key from base and params is appended, not replaced.
	assert.Equal(t, "/a?x=1&x=2", buildPath("/a?x=1", url.Values{"x": []string{"2"}}))
	assert.Equal(t, "/a?x=1&x=2", buildPath("/a", map[string]string{"x": "1"}, map[string]string{"x": "2"}))
}

func TestBuildPathTraversalNotCleaned(t *testing.T) {
	t.Parallel()

	// Unlike SafeRedirectPath, buildPath uses path.Join (not path.Clean) for
	// path params, so leading ".." segments are preserved.
	assert.Equal(t, "/a/../../etc/passwd", buildPath("/a", "../../etc/passwd"))
}

func TestBuildPathZeroValues(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/a?z=0", buildPath("/a", map[string]any{"z": 0}))
	assert.Equal(t, "/a?z=false", buildPath("/a", map[string]any{"z": false}))
	assert.Equal(t, "/a?z=%3Cnil%3E", buildPath("/a", map[string]any{"z": nil}))
}
