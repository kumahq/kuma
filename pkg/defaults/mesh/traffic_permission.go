package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var defaultTrafficPermissionResource = &core_mesh.TrafficPermissionResource{
	Spec: &mesh_proto.TrafficPermission{
		Sources: []*mesh_proto.Selector{
			{
				Match: map[string]string{
					mesh_proto.ServiceTag: "*",
				},
			},
		},
		Destinations: []*mesh_proto.Selector{
			{
				Match: map[string]string{
					mesh_proto.ServiceTag: "*",
				},
			},
		},
	},
}
