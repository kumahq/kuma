package once

import "sync"

type once struct {
	syncOnce sync.Once
	Value    any
	Err      error
}

func (o *once) Do(f func() (any, error)) {
	o.syncOnce.Do(func() {
		o.Value, o.Err = f()
	})
}

func newMap() *omap {
	return &omap{}
}

// omap is a concurrent map keyed by cache key, holding one *once per in-flight retrieval.
// It backs Cache.GetOrRetrieve's miss path, which every concurrent caller hits until the
// underlying value is cached, across every key (every mesh, every CLA, etc.) sharing the same
// Cache. A single mutex around map access here serializes all of those unrelated lookups on
// one lock, not just concurrent lookups of the same key; under load (hundreds of concurrent
// DataplaneWatchdog goroutines) that turned map access into the dominant cost of the request
// path. sync.Map is built for exactly this shape - a small, mostly-stable set of keys read far
// more often than written - and lets lookups for different keys proceed without contending on
// a shared lock.
type omap struct {
	m sync.Map // map[string]*once
}

func (c *omap) Get(key string) (*once, bool) {
	actual, loaded := c.m.LoadOrStore(key, &once{})
	return actual.(*once), !loaded
}

func (c *omap) Delete(key string) {
	c.m.Delete(key)
}
