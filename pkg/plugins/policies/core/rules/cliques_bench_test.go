package rules_test

import (
	"fmt"
	"testing"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_rules "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules"
	meshtrafficpermission_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
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
const benchMesh = "mesh-a"

var (
	// benchFromItems is the total number of `from` entries across all policies
	// matched to a single inbound, which is what BuildRules receives.
	benchFromItems = []int{50, 200, 500}
	benchInbounds  = []int{2, 6}
)

// benchMTPList builds one MeshTrafficPermission per `from` entry, each admitting a
// single distinct caller. tagsPerCaller controls how many tags identify a caller:
// 1 is the common `kuma.io/service` case, 2 adds a version dimension.
func benchMTPList(fromItems int, tagsPerCaller int) []core_model.Resource {
	items := make([]core_model.Resource, 0, fromItems)

	for i := 0; i < fromItems; i++ {
		tags := map[string]string{
			mesh_proto.ServiceTag: fmt.Sprintf("caller-%d", i),
		}
		if tagsPerCaller > 1 {
			tags["version"] = fmt.Sprintf("v%d", i%3)
		}

		items = append(items, builders.MeshTrafficPermission().
			WithMesh(benchMesh).
			WithName(fmt.Sprintf("mtp-%d", i)).
			WithTargetRef(common_api.TargetRef{Kind: common_api.Mesh}).
			AddFrom(common_api.TargetRef{
				Kind: common_api.MeshSubset,
				Tags: tags,
			}, meshtrafficpermission_api.Allow).
			Build())
	}

	return items
}

func benchByInbound(inbounds int, list []core_model.Resource) map[core_rules.InboundListener][]core_model.Resource {
	byInbound := map[core_rules.InboundListener][]core_model.Resource{}
	for i := 0; i < inbounds; i++ {
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
