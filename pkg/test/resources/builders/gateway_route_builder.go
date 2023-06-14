package builders

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type GatewayRouteBuilder struct {
	res *core_mesh.MeshGatewayRouteResource
}

func GatewayRoute() *GatewayRouteBuilder {
	return &GatewayRouteBuilder{
		res: &core_mesh.MeshGatewayRouteResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "gateway-route",
			},
			Spec: &mesh_proto.MeshGatewayRoute{},
		},
	}
}

func (gr *GatewayRouteBuilder) Build() *core_mesh.MeshGatewayRouteResource {
	if err := gr.res.Validate(); err != nil {
		panic(err)
	}
	return gr.res
}

func (gr *GatewayRouteBuilder) With(fn func(*core_mesh.MeshGatewayRouteResource)) *GatewayRouteBuilder {
	fn(gr.res)
	return gr
}

func (gr *GatewayRouteBuilder) WithName(name string) *GatewayRouteBuilder {
	gr.res.Meta.(*test_model.ResourceMeta).Name = name
	return gr
}

func (gr *GatewayRouteBuilder) WithGateway(gatewayName string) *GatewayRouteBuilder {
	gr.res.Spec.Selectors = []*mesh_proto.Selector{
		{
			Match: map[string]string{
				mesh_proto.ServiceTag: gatewayName,
			},
		},
	}
	return gr
}

func (gr *GatewayRouteBuilder) WithExactMatchHttpRoute(path string, backendServices ...string) *GatewayRouteBuilder {
	var backends []*mesh_proto.MeshGatewayRoute_Backend
	for _, backendService := range backendServices {
		backends = append(backends, &mesh_proto.MeshGatewayRoute_Backend{
			Weight: 1,
			Destination: map[string]string{
				"kuma.io/service": backendService,
			},
		})
	}
	gr.res.Spec.Conf = &mesh_proto.MeshGatewayRoute_Conf{
		Route: &mesh_proto.MeshGatewayRoute_Conf_Http{
			Http: &mesh_proto.MeshGatewayRoute_HttpRoute{
				Rules: []*mesh_proto.MeshGatewayRoute_HttpRoute_Rule{
					{
						Matches: []*mesh_proto.MeshGatewayRoute_HttpRoute_Match{
							{Path: &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
								Match: mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_EXACT,
								Value: path,
							}},
						},
						Backends: backends,
					},
				},
			},
		},
	}
	return gr
}

func (gr *GatewayRouteBuilder) WithTCPRoute(backend string) *GatewayRouteBuilder {
	gr.res.Spec.Conf = &mesh_proto.MeshGatewayRoute_Conf{
		Route: &mesh_proto.MeshGatewayRoute_Conf_Tcp{
			Tcp: &mesh_proto.MeshGatewayRoute_TcpRoute{
				Rules: []*mesh_proto.MeshGatewayRoute_TcpRoute_Rule{
					{
						Backends: []*mesh_proto.MeshGatewayRoute_Backend{
							{
								Weight: 1,
								Destination: map[string]string{
									"kuma.io/service": backend,
								},
							},
						},
					},
				},
			},
		},
	}
	return gr
}
