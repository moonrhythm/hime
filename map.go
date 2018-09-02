package hime

import "sync"

func cloneMap(m sync.Map) (r sync.Map) {
	m.Range(func(key, value interface{}) bool {
		r.Store(key, value)
		return true
	})
	return
}
