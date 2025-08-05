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

func (s *Sync[K, V]) LoadAndDelete(k K) (V, bool) {
	v, ok := s.inner.LoadAndDelete(k)
	if !ok {
		var zero V
		return zero, false
	}
	return v.(V), true
}

func (s *Sync[K, V]) Delete(k K) {
	s.inner.Delete(k)
}

func (s *Sync[K, V]) Swap(k K, v V) (V, bool) {
	prev, ok := s.inner.Swap(k, v)
	if !ok {
		var zero V
		return zero, false
	}
	return prev.(V), true
}

func (s *Sync[K, V]) CompareAndSwap(k K, old, n V) bool {
	return s.inner.CompareAndSwap(k, old, n)
}

func (s *Sync[K, V]) CompareAndDelete(k K, old V) bool {
	return s.inner.CompareAndDelete(k, old)
}

func (s *Sync[K, V]) Range(f func(k K, v V) bool) {
	s.inner.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}
