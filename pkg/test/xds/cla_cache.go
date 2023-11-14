package xds

import (
	"context"

	"google.golang.org/protobuf/proto"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
)

type DummyCLACache struct {
	OutboundTargets core_xds.EndpointMap
}

func (d *DummyCLACache) GetCLA(ctx context.Context, meshName, meshHash string, cluster envoy_common.Cluster, apiVersion core_xds.APIVersion, endpointMap core_xds.EndpointMap) (proto.Message, error) {
	return endpoints.CreateClusterLoadAssignment(cluster.Name(), d.OutboundTargets[cluster.Service()]), nil
}

var _ envoy_common.CLACache = &DummyCLACache{}
