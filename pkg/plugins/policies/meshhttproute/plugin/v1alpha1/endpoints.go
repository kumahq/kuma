package v1alpha1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/user"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

func generateEndpoints(
	proxy *core_xds.Proxy,
	ctx xds_context.Context,
	services envoy_common.Services,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	for _, serviceName := range services.Sorted() {
		// When no zone egress is present in a mesh Endpoints for ExternalServices
		// are specified in load assignment in DNS Cluster.
		// We are not allowed to add endpoints with DNS names through EDS.
		if !services[serviceName].HasExternalService() || ctx.Mesh.Resource.ZoneEgressEnabled() {
			for _, cluster := range services[serviceName].Clusters() {
				var endpoints core_xds.EndpointMap
				if cluster.Mesh() != "" {
					endpoints = ctx.Mesh.CrossMeshEndpoints[cluster.Mesh()]
				} else {
					endpoints = ctx.Mesh.EndpointMap
				}

				loadAssignment, err := ctx.ControlPlane.CLACache.GetCLA(user.Ctx(context.TODO(), user.ControlPlane), proxy.Dataplane.GetMeta().GetMesh(), ctx.Mesh.Hash, cluster, proxy.APIVersion, endpoints)
				if err != nil {
					return nil, errors.Wrapf(err, "could not get ClusterLoadAssignment for %s", serviceName)
				}

				resources.Add(&core_xds.Resource{
					Name:     cluster.Name(),
					Resource: loadAssignment,
				})
			}
		}
	}

	return resources, nil
}
