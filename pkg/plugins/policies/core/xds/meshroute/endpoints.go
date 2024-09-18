package meshroute

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/user"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func GenerateEndpoints(
	proxy *core_xds.Proxy,
	ctx xds_context.Context,
	services envoy_common.Services,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	for _, serviceName := range services.Sorted() {
		// When no zone egress is present in a mesh Endpoints for ExternalServices
		// are specified in load assignment in DNS Cluster.
		// We are not allowed to add endpoints with DNS names through EDS.
		service := services[serviceName]
		meshCtx := ctx.Mesh

		internalService := !ctx.Mesh.IsExternalService(serviceName)
		meshExternalService := isMeshExternalService(meshCtx.EndpointMap[serviceName])
		externalServiceThroughEgress := ctx.Mesh.IsExternalService(serviceName) && !meshExternalService && meshCtx.Resource.ZoneEgressEnabled()
		if internalService || meshExternalService || externalServiceThroughEgress {
			for _, cluster := range service.Clusters() {
				var endpoints core_xds.EndpointMap
				if cluster.Mesh() != "" {
					endpoints = meshCtx.CrossMeshEndpoints[cluster.Mesh()]
				} else {
					endpoints = meshCtx.EndpointMap
				}

				loadAssignment, err := ctx.ControlPlane.CLACache.GetCLA(
					user.Ctx(context.TODO(), user.ControlPlane),
					proxy.Dataplane.GetMeta().GetMesh(),
					meshCtx.Hash,
					cluster,
					proxy.APIVersion,
					endpoints,
				)
				if err != nil {
					return nil, errors.Wrapf(err,
						"could not get ClusterLoadAssignment for %s",
						serviceName,
					)
				}

				resources.Add(&core_xds.Resource{
					Name:     cluster.Name(),
					Origin:   generator.OriginOutbound,
					Resource: loadAssignment,
				})
			}
		}
	}

	return resources, nil
}
