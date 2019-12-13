package policy

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

// ConnectionPolicy is a Policy that is applied on a connection between two data planes that match source and destination.
type ConnectionPolicy interface {
	core_model.Resource
	Sources() []*mesh_proto.Selector
	Destinations() []*mesh_proto.Selector
}

// ConnectionPolicyMap holds the most specific ConnectionPolicy for each outbound interface of a Dataplane.
type ConnectionPolicyMap map[core_xds.ServiceName]ConnectionPolicy
