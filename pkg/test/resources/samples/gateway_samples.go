package samples

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

func BackendGatewayRoute() *core_mesh.MeshGatewayRouteResource {
	return builders.GatewayRoute().
		WithName("sample-gateway-route").
		WithGateway("sample-gateway").
		WithExactMatchHttpRoute("/", "backend").
		Build()
}

func BackendGatewayTCPRoute() *core_mesh.MeshGatewayRouteResource {
	return builders.GatewayRoute().
		WithName("sample-tcp-gateway-route").
		WithGateway("sample-gateway").
		WithTCPRoute("backend").
		Build()
}

func GatewayResource() *core_mesh.MeshGatewayResource {
	return &core_mesh.MeshGatewayResource{
		Meta: &test_model.ResourceMeta{Name: "sample-gateway", Mesh: "default"},
		Spec: &mesh_proto.MeshGateway{
			Selectors: []*mesh_proto.Selector{
				{
					Match: map[string]string{
						mesh_proto.ServiceTag: "sample-gateway",
					},
				},
			},
			Conf: &mesh_proto.MeshGateway_Conf{
				Listeners: []*mesh_proto.MeshGateway_Listener{
					{
						Protocol: mesh_proto.MeshGateway_Listener_HTTP,
						Port:     8080,
					},
				},
			},
		},
	}
}

func GatewayTCPResource() *core_mesh.MeshGatewayResource {
	return &core_mesh.MeshGatewayResource{
		Meta: &test_model.ResourceMeta{Name: "sample-gateway", Mesh: "default"},
		Spec: &mesh_proto.MeshGateway{
			Selectors: []*mesh_proto.Selector{
				{
					Match: map[string]string{
						mesh_proto.ServiceTag: "sample-gateway",
					},
				},
			},
			Conf: &mesh_proto.MeshGateway_Conf{
				Listeners: []*mesh_proto.MeshGateway_Listener{
					{
						Protocol: mesh_proto.MeshGateway_Listener_TCP,
						Port:     8080,
					},
				},
			},
		},
	}
}
