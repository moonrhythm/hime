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

func getBytes() *bytes.Buffer {
	b := bytesPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func putBytes(b *bytes.Buffer) {
	b.Reset()
	bytesPool.Put(b)
}
