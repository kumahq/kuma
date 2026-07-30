package v1alpha1

import (
	"testing"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

func TestBackendRefToClusterNameForRouteUsesResolvedRequestMirrorBackendRefs(t *testing.T) {
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
		Build().
		Mesh

	mirrorBackendRef := common_api.BackendRef{
		TargetRef: common_api.TargetRef{
			Kind:      common_api.MeshService,
			Name:      pointer.To("payments"),
			Namespace: pointer.To("kuma-demo"),
		},
	}

	route := requestMirrorRoute("kuma-demo", mirrorBackendRef)

	clusterCache := map[common_api.BackendRefHash]string{}
	servicesAcc := envoy_common.NewServicesAccumulator(meshCtx.GetTLSReadiness())

	backendRefToClusterName := backendRefToClusterNameForRoute(clusterCache, servicesAcc, route, meshCtx)

	if got, want := backendRefToClusterName[mirrorBackendRef.Hash()], kri.From(payments).String(); got != want {
		t.Fatalf("request-mirror backendRef hash %q resolved to cluster %q, want %q", mirrorBackendRef.Hash(), got, want)
	}
}

func TestBackendRefToClusterNameForRouteScopesRequestMirrorBackendRefAliasesByOriginNamespace(t *testing.T) {
	t.Parallel()

	paymentsTeamA := builders.MeshService().
		WithName("payments-team-a-hash").
		WithMesh("default").
		WithLabels(map[string]string{
			mesh_proto.DisplayName:      "payments",
			mesh_proto.KubeNamespaceTag: "team-a",
		}).
		AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
		Build()

	paymentsTeamB := builders.MeshService().
		WithName("payments-team-b-hash").
		WithMesh("default").
		WithLabels(map[string]string{
			mesh_proto.DisplayName:      "payments",
			mesh_proto.KubeNamespaceTag: "team-b",
		}).
		AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
		Build()

	meshCtx := xds_builders.Context().
		WithMeshLocalResources([]core_model.Resource{paymentsTeamA, paymentsTeamB}).
		Build().
		Mesh

	mirrorBackendRef := common_api.BackendRef{
		TargetRef: common_api.TargetRef{
			Kind: common_api.MeshService,
			Name: pointer.To("payments"),
		},
	}

	clusterCache := map[common_api.BackendRefHash]string{}
	servicesAcc := envoy_common.NewServicesAccumulator(meshCtx.GetTLSReadiness())

	teamABackendRefToClusterName := backendRefToClusterNameForRoute(clusterCache, servicesAcc, requestMirrorRoute("team-a", mirrorBackendRef), meshCtx)
	teamBBackendRefToClusterName := backendRefToClusterNameForRoute(clusterCache, servicesAcc, requestMirrorRoute("team-b", mirrorBackendRef), meshCtx)

	hash := mirrorBackendRef.Hash()
	if got, want := teamABackendRefToClusterName[hash], kri.From(paymentsTeamA).String(); got != want {
		t.Fatalf("team-a request-mirror backendRef hash %q resolved to cluster %q, want %q", hash, got, want)
	}
	if got, want := teamBBackendRefToClusterName[hash], kri.From(paymentsTeamB).String(); got != want {
		t.Fatalf("team-b request-mirror backendRef hash %q resolved to cluster %q, want %q", hash, got, want)
	}
}

func requestMirrorRoute(namespace string, mirrorBackendRef common_api.BackendRef) api.Route {
	return api.Route{
		Origin: kri.Identifier{
			ResourceType: api.MeshHTTPRouteType,
			Mesh:         core_model.DefaultMesh,
			Namespace:    namespace,
			Name:         "web-route",
		},
		Filters: []api.Filter{{
			Type: api.RequestMirrorType,
			RequestMirror: &api.RequestMirror{
				BackendRef: mirrorBackendRef,
			},
		}},
	}
}
