package hime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPath(t *testing.T) {
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
