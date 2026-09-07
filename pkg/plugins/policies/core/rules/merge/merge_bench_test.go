package merge_test

import (
	"fmt"
	"testing"
	"time"

	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/merge"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func benchmarkConfs(b *testing.B, n int) {
	b.Helper()
	confs := []any{}
	for i := range n {
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
	for i := range 40 {
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

func benchmarkMergeByKey(b *testing.B, policies, rulesPerPolicy int) {
	b.Helper()
	confs := []any{}
	for i := range policies {
		rules := []meshhttproute_api.Rule{}
		for j := range rulesPerPolicy {
			rules = append(rules, meshhttproute_api.Rule{
				Matches: []meshhttproute_api.Match{{
					Path: &meshhttproute_api.PathMatch{
						Type:  meshhttproute_api.Exact,
						Value: fmt.Sprintf("/policy-%d/rule-%d", i, j),
					},
					Method: pointer.To(meshhttproute_api.Method("GET")),
				}},
			})
		}
		confs = append(confs, meshhttproute_api.PolicyDefault{Rules: rules})
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

func BenchmarkMergeByKey10x10(b *testing.B) { benchmarkMergeByKey(b, 10, 10) }
func BenchmarkMergeByKey20x20(b *testing.B) { benchmarkMergeByKey(b, 20, 20) }
func BenchmarkMergeByKey40x10(b *testing.B) { benchmarkMergeByKey(b, 40, 10) }
