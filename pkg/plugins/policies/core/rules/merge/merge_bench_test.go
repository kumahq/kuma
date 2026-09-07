package merge_test

import (
	"fmt"
	"testing"
	"time"

	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/merge"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func benchmarkConfs(b *testing.B, n int) {
	b.Helper()
	confs := []any{}
	for i := 0; i < n; i++ {
		confs = append(confs, meshtimeout_api.Conf{
			ConnectionTimeout: pointer.To(k8s.Duration{Duration: time.Duration(i+1) * time.Second}),
			IdleTimeout:       pointer.To(k8s.Duration{Duration: time.Duration(20+i%7) * time.Second}),
			Http: &meshtimeout_api.Http{
				RequestTimeout:    pointer.To(k8s.Duration{Duration: 15 * time.Second}),
				StreamIdleTimeout: pointer.To(k8s.Duration{Duration: time.Duration(i+1) * time.Minute}),
			},
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := merge.Confs(confs)
		if err != nil {
			b.Fatal(err)
		}
		if len(out) != 1 {
			b.Fatalf("expected 1 merged conf, got %d", len(out))
		}
	}
}

func BenchmarkConfs1(b *testing.B)  { benchmarkConfs(b, 1) }
func BenchmarkConfs2(b *testing.B)  { benchmarkConfs(b, 2) }
func BenchmarkConfs10(b *testing.B) { benchmarkConfs(b, 10) }
func BenchmarkConfs40(b *testing.B) { benchmarkConfs(b, 40) }

func BenchmarkConfsMeshWidePolicies(b *testing.B) {
	confs := []any{}
	for i := 0; i < 40; i++ {
		confs = append(confs, meshtimeout_api.Conf{
			ConnectionTimeout: pointer.To(k8s.Duration{Duration: time.Duration(i+1) * time.Second}),
			Http: &meshtimeout_api.Http{
				RequestTimeout: pointer.To(k8s.Duration{Duration: 15 * time.Second}),
			},
		})
	}
	b.Run(fmt.Sprintf("n=%d", len(confs)), func(b *testing.B) {
		benchmarkConfs(b, len(confs))
	})
}
