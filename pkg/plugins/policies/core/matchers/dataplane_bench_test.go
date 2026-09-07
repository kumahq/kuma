package matchers_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

func benchResources(b *testing.B, numPolicies int) xds_context.Resources {
	b.Helper()
	core_plugins.InitAll(policies.NameToModule)
	resources := xds_context.NewResources()
	list := &meshtimeout_api.MeshTimeoutResourceList{}
	for i := range numPolicies {
		var targetRef common_api.TargetRef
		switch i % 3 {
		case 0:
			targetRef = common_api.TargetRef{Kind: common_api.Mesh}
		case 1:
			targetRef = common_api.TargetRef{
				Kind:   common_api.Dataplane,
				Labels: pointer.To(map[string]string{"app": "backend"}),
			}
		default:
			targetRef = common_api.TargetRef{
				Kind:   common_api.Dataplane,
				Labels: pointer.To(map[string]string{"team": "infra"}),
			}
		}
		mt := builders.MeshTimeout().
			WithName(fmt.Sprintf("mt-%d", i)).
			WithTargetRef(targetRef).
			AddRule([]common_api.Match{}, meshtimeout_api.Conf{
				ConnectionTimeout: pointer.To(k8s.Duration{Duration: 10 * time.Second}),
			}).
			Build()
		if err := list.AddItem(mt); err != nil {
			b.Fatal(err)
		}
	}
	resources.MeshLocalResources[meshtimeout_api.MeshTimeoutType] = list
	return resources
}

func benchDataplane(b *testing.B) *core_mesh.DataplaneResource {
	b.Helper()
	return builders.Dataplane().
		WithName("dp-1").
		WithMesh("default").
		WithVersion("1").
		WithLabels(map[string]string{
			mesh_proto.ZoneTag:          "zone-1",
			mesh_proto.KubeNamespaceTag: "default",
			"team":                      "infra",
			"app":                       "backend",
		}).
		WithAddress("127.0.0.1").
		WithServices("backend").
		WithInboundOfTagsAndProtocol("http", "kuma.io/display-name", "web").
		WithTransparentProxying(15001, 15006, "").
		Build()
}

func benchCache() *matchers.PolicyMatchingCache {
	return matchers.NewPolicyMatchingCache(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bench_policy_matching_cache",
	}, []string{"result"}), 1000)
}

func benchmarkPluginLoopCacheHit(b *testing.B, hoistHash bool) {
	b.Helper()
	dpp := benchDataplane(b)
	resources := benchResources(b, 10)
	cache := benchCache()
	opts := []core_plugins.MatchedPoliciesOption{core_plugins.WithCache(cache, "mesh-hash")}
	if hoistHash {
		opts = append(opts, core_plugins.WithDataplaneHash(dpp.Hash()))
	}

	if _, err := matchers.MatchedPolicies(meshtimeout_api.MeshTimeoutType, dpp, resources, opts...); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for range 17 {
			if _, err := matchers.MatchedPolicies(meshtimeout_api.MeshTimeoutType, dpp, resources, opts...); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkMatchedPoliciesPluginLoopCacheHit(b *testing.B) {
	benchmarkPluginLoopCacheHit(b, false)
}

func BenchmarkMatchedPoliciesPluginLoopCacheHitHoistedHash(b *testing.B) {
	benchmarkPluginLoopCacheHit(b, true)
}

func BenchmarkMatchedPoliciesCacheMiss(b *testing.B) {
	dpp := benchDataplane(b)
	resources := benchResources(b, 10)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := matchers.MatchedPolicies(meshtimeout_api.MeshTimeoutType, dpp, resources); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildCacheKey(b *testing.B) {
	dpp := benchDataplane(b)
	cfg := core_plugins.NewMatchedPoliciesConfig(core_plugins.WithCache(nil, "mesh-hash"))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = matchers.BuildCacheKey(string(meshtimeout_api.MeshTimeoutType), cfg, dpp)
	}
}
