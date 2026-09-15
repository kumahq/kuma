package maps

import "sync"

// Sync is a simple wrapper around sync.Map that provides type-safe methods
type Sync[K, V any] struct {
	inner sync.Map
}

func (s *Sync[K, V]) Load(k K) (V, bool) {
	v, ok := s.inner.Load(k)
	if !ok {
		var zero V
		return zero, false
	}
	return v.(V), true
}

func (s *Sync[K, V]) Store(k K, v V) {
	s.inner.Store(k, v)
}

func (s *Sync[K, V]) LoadOrStore(k K, store V) (V, bool) {
	v, ok := s.inner.LoadOrStore(k, store)
	return v.(V), ok
}

func (s *Sync[K, V]) Delete(k K) {
	s.inner.Delete(k)
}
