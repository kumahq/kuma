package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var defaultTrafficRouteResource = &core_mesh.TrafficRouteResource{
	Spec: &mesh_proto.TrafficRoute{
		Sources: []*mesh_proto.Selector{{
			Match: mesh_proto.MatchAnyService(),
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: mesh_proto.MatchAnyService(),
		}},
		Conf: &mesh_proto.TrafficRoute_Conf{
			Destination: mesh_proto.MatchAnyService(),
			LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
				LbType: &mesh_proto.TrafficRoute_LoadBalancer_RoundRobin_{},
			},
		},
	},
}
