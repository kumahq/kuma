package v1alpha1

import (
	"testing"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	meshroute_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	test_policies "github.com/kumahq/kuma/v3/pkg/test/policies"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

func TestGenerateFromServiceUsesResolvedRequestMirrorBackendRefs(t *testing.T) {
	t.Parallel()

	payments := builders.MeshService().
		WithName("payments-hash").
		WithMesh("default").
		WithLabels(map[string]string{
			mesh_proto.DisplayName:      "payments",
			mesh_proto.KubeNamespaceTag: "kuma-demo",
		}).
		AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
		Build()

	meshCtx := xds_builders.Context().
		WithMeshLocalResources([]core_model.Resource{payments}).
		AddServiceProtocol("backend", core_meta.ProtocolHTTP).
		Build().
		Mesh

	proxy := xds_builders.Proxy().
		WithDataplane(builders.Dataplane().
			WithName("web-01").
			WithAddress("192.168.0.2").
			WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http"),
		).
		WithOutbounds(xds_types.Outbounds{
			{
				LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Port: 10001,
					Tags: map[string]string{mesh_proto.ServiceTag: "backend"},
				},
			},
		}).
		Build()

	svc := meshroute_xds.DestinationService{
		Outbound:            proxy.Outbounds[0],
		Protocol:            core_meta.ProtocolHTTP,
		KumaServiceTagValue: "backend",
	}

	mirrorBackendRef := common_api.BackendRef{
		TargetRef: common_api.TargetRef{
			Kind:      common_api.MeshService,
			Name:      pointer.To("payments"),
			Namespace: pointer.To("kuma-demo"),
		},
	}

	rules := core_rules.ToRules{
		Rules: core_rules.Rules{
			test_policies.NewRule(subsetutils.MeshService("backend"), api.PolicyDefault{
				Rules: []api.Rule{{
					Matches: []api.Match{{Path: &api.PathMatch{Type: api.PathPrefix, Value: "/"}}},
					Default: api.RuleConf{
						Filters: &[]api.Filter{{
							Type: api.RequestMirrorType,
							RequestMirror: &api.RequestMirror{
								BackendRef: mirrorBackendRef,
							},
						}},
					},
				}},
			}),
		},
	}

	clusterCache := map[common_api.BackendRefHash]string{}
	servicesAcc := envoy_common.NewServicesAccumulator(meshCtx.GetTLSReadiness())

	if _, err := generateFromService(meshCtx, proxy, clusterCache, servicesAcc, rules, svc); err != nil {
		t.Fatalf("generateFromService() error = %v", err)
	}

	if _, found := clusterCache[mirrorBackendRef.Hash()]; !found {
		t.Fatalf("expected request-mirror backendRef hash %q to resolve to a cluster", mirrorBackendRef.Hash())
	}
}
