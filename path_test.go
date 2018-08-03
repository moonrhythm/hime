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
		{"https://google.com", "https://google.com"},
		{"https://google.com/test", "https://google.com/test"},
		{"https://google.com/test?p=1", "https://google.com/test?p=1"},
		{"http://google.com/test?p=1", "http://google.com/test?p=1"},
		{"//a?p=1", "//a?p=1"},
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
		Params []interface{}
		Output string
	}{
		{"", []interface{}{""}, "/"},
		{"/", []interface{}{"/"}, "/"},
		{"/a", []interface{}{}, "/a"},
		{"/a", []interface{}{"/b"}, "/a/b"},
		{"/a/", []interface{}{"/b/", "/c/"}, "/a/b/c"},
		{"/a", []interface{}{url.Values{"id": []string{"10"}}}, "/a?id=10"},
		{"/a", []interface{}{"/b", url.Values{"id": []string{"10"}}}, "/a/b?id=10"},
		{"/a", []interface{}{"/b/", url.Values{"id": []string{"10"}}}, "/a/b?id=10"},
		{"/a", []interface{}{"/b/", map[string]string{"id": "10"}}, "/a/b?id=10"},
		{"/a", []interface{}{"/b/", map[string]interface{}{"id": 10}}, "/a/b?id=10"},
		{"/a", []interface{}{"/b", &Param{Name: "id", Value: 3456}}, "/a/b?id=3456"},
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
