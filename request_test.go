package hime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveComma(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input  string
		Output string
	}{
		{"", ""},
		{"123", "123"},
		{"12,345", "12345"},
		{" 12, ,, 34,5, ,,", " 12  345 "},
		{"12,345.67", "12345.67"},
	}

	for _, c := range cases {
		assert.Equal(t, c.Output, removeComma(c.Input))
	}
}
