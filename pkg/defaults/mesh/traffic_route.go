package mesh

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
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

// TrafficRoute needs to contain mesh name inside it. Otherwise if the name is the same (ex. "allow-all") creating new mesh would fail because there is already resource of name "allow-all" which is unique key on K8S
func defaultTrafficRouteKey(meshName string) model.ResourceKey {
	return model.ResourceKey{
		Mesh: meshName,
		Name: fmt.Sprintf("route-all-%s", meshName),
	}
}
