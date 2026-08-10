package matchers_test

import (
	"fmt"
	"sync"
	"testing"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/v2/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/matchers"
	meshtrafficpermission_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

// Benchmarks don't go through test.RunSpecs, which is what normally registers the
// policy plugins, so the resource types have to be registered here.
var initPolicyPlugins = sync.OnceFunc(func() {
	core_plugins.InitAll(policies.NameToModule)
})

// This benchmark reproduces the shape of data seen on a large zone control plane
// where Kubernetes Services in a namespace use selectors that are wider than the
// Deployment they front. Every Service then selects every pod in the namespace, so
// each Dataplane ends up with one inbound per Service rather than one per container
// port. The inbounds that do not correspond to a real port on the pod stay in the
// Ignored state, and since Ignored inbounds are matched like any other, the cost of
// MatchedPolicies grows with the inflated inbound count.
//
// All names, namespaces and identities below are synthetic.
const (
	benchMesh      = "mesh-a"
	benchNamespace = "ns-a"
	benchZone      = "zone-a"
)

// benchInboundShape describes how many inbounds a Dataplane carries.
// The tight variant is what a correctly scoped Service selector produces: one
// inbound per real container port. The wide variant is what a namespace-wide
// selector produces: the same real ports plus one Ignored inbound per unrelated
// Service in the namespace.
type benchInboundShape struct {
	name    string
	ready   int
	ignored int
}

var benchInboundShapes = []benchInboundShape{
	{name: "inbounds2_tight_selectors", ready: 2, ignored: 0},
	{name: "inbounds6_wide_selectors", ready: 2, ignored: 4},
}

var benchPolicyCounts = []int{100, 500, 2000}

// benchDataplane builds a Dataplane with `ready` Ready inbounds and `ignored`
// Ignored inbounds, mimicking a pod picked up by more Services than it has ports.
func benchDataplane(shape benchInboundShape) *core_mesh.DataplaneResource {
	inbounds := make([]*mesh_proto.Dataplane_Networking_Inbound, 0, shape.ready+shape.ignored)
	port := uint32(8080)
	for range shape.ready {
		inbounds = append(inbounds, &mesh_proto.Dataplane_Networking_Inbound{
			Port:        port,
			Name:        fmt.Sprintf("port-%d", port),
			State:       mesh_proto.Dataplane_Networking_Inbound_Ready,
			ServicePort: port,
		})
		port++
	}
	for range shape.ignored {
		inbounds = append(inbounds, &mesh_proto.Dataplane_Networking_Inbound{
			Port:        port,
			Name:        fmt.Sprintf("port-%d", port),
			State:       mesh_proto.Dataplane_Networking_Inbound_Ignored,
			ServicePort: port,
		})
		port++
	}

	return &core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Mesh: benchMesh,
			Name: "app-a-6bf4654569-cxq49",
			Labels: map[string]string{
				mesh_proto.KubeNamespaceTag:    benchNamespace,
				mesh_proto.ZoneTag:             benchZone,
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				"app":                          "app-a",
			},
		},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "10.0.0.1",
				Inbound: inbounds,
			},
		},
	}
}

// benchPolicies builds `n` MeshTrafficPermissions in the shape a per-service
// dependency allow-list produces: mostly Dataplane-scoped policies that name a
// single workload, plus a small number of mesh-wide baseline policies.
func benchPolicies(n int) xds_context.Resources {
	items := make([]*meshtrafficpermission_api.MeshTrafficPermissionResource, 0, n)
	for i := range n {
		var targetRef *common_api.TargetRef
		switch {
		case i%50 == 0:
			// mesh-wide baseline policy
			targetRef = &common_api.TargetRef{Kind: common_api.Mesh}
		case i%10 == 1:
			// policy that selects the Dataplane under test
			targetRef = &common_api.TargetRef{
				Kind:   common_api.Dataplane,
				Labels: pointer.To(map[string]string{"app": "app-a"}),
			}
		default:
			// policy that selects some other workload in the mesh
			targetRef = &common_api.TargetRef{
				Kind:   common_api.Dataplane,
				Labels: pointer.To(map[string]string{"app": fmt.Sprintf("app-%d", i)}),
			}
		}

		items = append(items, &meshtrafficpermission_api.MeshTrafficPermissionResource{
			Meta: &test_model.ResourceMeta{
				Mesh: benchMesh,
				Name: fmt.Sprintf("mtp-%d", i),
				Labels: map[string]string{
					mesh_proto.ZoneTag: benchZone,
				},
			},
			Spec: &meshtrafficpermission_api.MeshTrafficPermission{
				TargetRef: targetRef,
				Rules: &[]meshtrafficpermission_api.Rule{{
					Default: meshtrafficpermission_api.RuleConf{
						Allow: &[]common_api.Match{{
							SpiffeID: &common_api.SpiffeIDMatch{
								Type:  common_api.ExactMatchType,
								Value: fmt.Sprintf("spiffe://%s/ns/%s/sa/caller-%d", benchMesh, benchNamespace, i),
							},
						}},
					},
				}},
			},
		})
	}

	return xds_context.Resources{
		MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
			meshtrafficpermission_api.MeshTrafficPermissionType: &meshtrafficpermission_api.MeshTrafficPermissionResourceList{
				Items: items,
			},
		},
	}
}

func BenchmarkMatchedPoliciesInboundFanout(b *testing.B) {
	initPolicyPlugins()

	for _, shape := range benchInboundShapes {
		for _, policyCount := range benchPolicyCounts {
			dpp := benchDataplane(shape)
			resources := benchPolicies(policyCount)

			b.Run(fmt.Sprintf("%s/policies%d", shape.name, policyCount), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, err := matchers.MatchedPolicies(
						meshtrafficpermission_api.MeshTrafficPermissionType,
						dpp,
						resources,
					); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// BenchmarkDppSelectedByPolicy isolates the per-policy selection step that
// MatchedPolicies calls once per policy, so the cost attributable to the inbound
// count alone is visible without the surrounding rule building.
func BenchmarkDppSelectedByPolicy(b *testing.B) {
	initPolicyPlugins()

	meshRef := common_api.TargetRef{Kind: common_api.Mesh}

	for _, shape := range benchInboundShapes {
		dpp := benchDataplane(shape)
		meta := &test_model.ResourceMeta{Mesh: benchMesh, Name: "mtp-mesh-wide"}

		b.Run(shape.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, _, _, err := matchers.DppSelectedByPolicy(meta, meshRef, dpp, nil, xds_context.Resources{}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
