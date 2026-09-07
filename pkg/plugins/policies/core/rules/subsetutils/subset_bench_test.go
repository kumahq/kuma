package subsetutils_test

import (
	"fmt"
	"testing"

	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
)

func BenchmarkIsSubset(b *testing.B) {
	newLabels := func(n int) map[string]string {
		m := map[string]string{}
		for i := range n {
			m[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("value-%d", i)
		}
		return m
	}
	for _, n := range []int{1, 5, 10} {
		query := subsetutils.NewSubset(newLabels(n))
		match := subsetutils.NewSubset(newLabels(n))
		mismatch := subsetutils.NewSubset(map[string]string{"app": "demo"})
		b.Run(fmt.Sprintf("labels=%d/match", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if !query.IsSubset(match) {
					b.Fatal("expected subset")
				}
			}
		})
		b.Run(fmt.Sprintf("labels=%d/mismatch", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if query.IsSubset(mismatch) {
					b.Fatal("expected no subset")
				}
			}
		})
	}
}
