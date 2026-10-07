package once

import (
	"fmt"
	"testing"
)

// BenchmarkOmapGetConcurrentKeys simulates the shape reported in
// https://github.com/kumahq/kuma/issues/16188: many goroutines (e.g. one per
// DataplaneWatchdog) concurrently calling Get across a small, stable set of distinct keys
// (e.g. one per mesh). Before this change, every one of these calls contended on a single
// mutex regardless of which key it was for.
func BenchmarkOmapGetConcurrentKeys(b *testing.B) {
	const keyCount = 50

	m := newMap()
	for i := range keyCount {
		m.Get(fmt.Sprintf("key-%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(fmt.Sprintf("key-%d", i%keyCount))
			i++
		}
	})
}
