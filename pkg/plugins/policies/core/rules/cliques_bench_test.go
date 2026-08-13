package rules_test

import (
	"fmt"
	"sync"
	"testing"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/v2/pkg/core/plugins"
	core_apis "github.com/kumahq/kuma/v2/pkg/core/resources/apis"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies"
	core_rules "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules"
	meshtrafficpermission_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

// These benchmarks exercise the legacy `from`-list rule builder that this branch
// uses for MeshTrafficPermission. BuildFromRules flattens the `from` entries of
// every matched policy into one list per inbound listener, then BuildRules turns
// that list into an intersection graph, enumerates maximal cliques
// (BronKerbosch), and for every tag combination of every clique calls createRule
// over the whole item list.
//
// The shape modeled here is a per-service dependency allow-list: each policy
// names the callers it admits with `from[].targetRef.kind: MeshSubset`, which is
// how a zone's MeshTrafficPermissions look when they are generated per service.
// The inbound counts mirror a pod fronted by correctly scoped Service selectors
// (2) versus namespace-wide ones (6).
//
// All names and identities are synthetic.
const (
	benchMesh = "mesh-a"
	benchZone = "zone-a"
)

// Benchmarks don't go through test.RunSpecs, which is what normally registers the
// policy plugins, so the resource types have to be registered here.
var initPolicyPlugins = sync.OnceFunc(func() {
	core_plugins.InitAll(core_apis.NameToModule)
	core_plugins.InitAll(policies.NameToModule)
})

var (
	// benchFromItems is the total number of `from` entries across all policies
	// matched to a single inbound, which is what BuildRules receives.
	benchFromItems = []int{50, 200, 500}
	benchInbounds  = []int{2, 6}
)

// benchMTPList builds one MeshTrafficPermission per `from` entry, each admitting a
// single distinct caller. tagsPerCaller controls how many tags identify a caller:
// 1 is the common `kuma.io/service` case, 2 adds a version dimension.
func benchMTPList(fromItems int, tagsPerCaller int) core_model.ResourceList {
	items := make([]*meshtrafficpermission_api.MeshTrafficPermissionResource, 0, fromItems)

	for i := range fromItems {
		tags := map[string]string{
			mesh_proto.ServiceTag: fmt.Sprintf("caller-%d", i),
		}
		if tagsPerCaller > 1 {
			tags["version"] = fmt.Sprintf("v%d", i%3)
		}

		res := builders.MeshTrafficPermission().
			WithMesh(benchMesh).
			WithName(fmt.Sprintf("mtp-%d", i)).
			WithTargetRef(common_api.TargetRef{Kind: common_api.Mesh}).
			AddFrom(common_api.TargetRef{
				Kind: common_api.MeshSubset,
				Tags: pointer.To(tags),
			}, meshtrafficpermission_api.Allow).
			Build()
		// The builder has no setter for labels, and matching needs the zone.
		res.Meta.(*test_model.ResourceMeta).Labels = map[string]string{mesh_proto.ZoneTag: benchZone}

		items = append(items, res)
	}

	return &meshtrafficpermission_api.MeshTrafficPermissionResourceList{Items: items}
}

func benchByInbound(inbounds int, list core_model.ResourceList) map[core_rules.InboundListener]core_model.ResourceList {
	byInbound := map[core_rules.InboundListener]core_model.ResourceList{}
	for i := range inbounds {
		byInbound[core_rules.InboundListener{
			Address: "10.0.0.1",
			Port:    uint32(8080 + i),
		}] = list
	}
	return byInbound
}

// BenchmarkBuildFromRulesCliques measures the production entry point: the whole
// per-inbound loop, including the clique enumeration and createRule expansion.
func BenchmarkBuildFromRulesCliques(b *testing.B) {
	initPolicyPlugins()

	for _, tagsPerCaller := range []int{1, 2} {
		for _, inbounds := range benchInbounds {
			for _, fromItems := range benchFromItems {
				list := benchMTPList(fromItems, tagsPerCaller)
				byInbound := benchByInbound(inbounds, list)

				name := fmt.Sprintf("tags%d/inbounds%d/from%d", tagsPerCaller, inbounds, fromItems)
				b.Run(name, func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if _, err := core_rules.BuildFromRules(byInbound); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}
