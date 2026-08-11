package meshroute

import (
	"context"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/pkg/core/user"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

// GenerateEndpoints creates a ClusterLoadAssignment for every cluster in
// generatedClusters. Clusters that GenerateClusters skipped - a
// MeshExternalService with no zone egress to reach it through, for example -
// are skipped here too: an endpoint resource that no cluster references makes
// the whole snapshot inconsistent, and the proxy is then sent nothing at all.
func GenerateEndpoints(
	proxy *core_xds.Proxy,
	ctx xds_context.Context,
	services envoy_common.Services,
	generatedClusters *core_xds.ResourceSet,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	clusterResources := generatedClusters.Resources(envoy_resource.ClusterType)

	for _, serviceName := range services.Sorted() {
		service := services[serviceName]
		meshCtx := ctx.Mesh

		for _, cluster := range service.Clusters() {
			if _, ok := clusterResources[cluster.Name()]; !ok {
				continue
			}

			loadAssignment, err := ctx.ControlPlane.CLACache.GetCLA(
				user.Ctx(context.TODO(), user.ControlPlane),
				proxy.Dataplane.GetMeta().GetMesh(),
				meshCtx.Hash,
				cluster,
				proxy.APIVersion,
				meshCtx.EndpointMap,
			)
			if err != nil {
				return nil, errors.Wrapf(err,
					"could not get ClusterLoadAssignment for %s",
					serviceName,
				)
			}

			resources.Add(&core_xds.Resource{
				Name:           cluster.Name(),
				Origin:         metadata.OriginOutbound,
				Resource:       loadAssignment,
				ResourceOrigin: service.BackendRef().Resource(),
			})
		}
	}

	return resources, nil
}
