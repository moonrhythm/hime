package hime

import (
	"bytes"
	"sync"
)

var bytesPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}
