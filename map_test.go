package hime

import "sync"

func mapLen(m *sync.Map) (i int) {
	m.Range(func(_, _ any) bool {
		i++
		return true
	})
	return
}
