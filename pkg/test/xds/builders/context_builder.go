package builders

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/test/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

type ContextBuilder struct {
	res *xds_context.Context
}

func Context() *ContextBuilder {
	return &ContextBuilder{
		res: &xds_context.Context{
			Mesh: xds_context.MeshContext{
				Resource:                  samples.MeshDefault(),
				EndpointMap:               map[core_xds.ServiceName][]core_xds.Endpoint{},
				ServicesInformation:       map[string]*xds_context.ServiceInformation{},
				MeshServiceIdentity:       map[string]topology.MeshServiceIdentity{},
				MeshExternalServiceByName: map[string]*meshexternalservice_api.MeshExternalServiceResource{},
			},
			ControlPlane: &xds_context.ControlPlaneContext{
				CLACache: &xds.DummyCLACache{OutboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{}},
				Zone:     "test-zone",
				Secrets:  &xds.TestSecrets{},
			},
		},
	}
}

func (mc *ContextBuilder) Build() *xds_context.Context {
	mc.res.Mesh.MeshServiceIdentity = topology.BuildMeshServiceIdentityMap(
		mc.res.Mesh.Resources.MeshServices().Items, mc.res.Mesh.EndpointMap,
	)
	for _, mes := range mc.res.Mesh.Resources.MeshExternalServices().Items {
		mc.res.Mesh.MeshExternalServiceByName[mes.GetMeta().GetName()] = mes
	}
	return mc.res
}

func (mc *ContextBuilder) With(fn func(*xds_context.Context)) *ContextBuilder {
	fn(mc.res)
	return mc
}

func (mc *ContextBuilder) WithEndpointMap(endpointMap *EndpointMapBuilder) *ContextBuilder {
	mc.res.Mesh.EndpointMap = endpointMap.Build()
	mc.res.ControlPlane.CLACache.(*xds.DummyCLACache).OutboundTargets = endpointMap.Build()
	return mc
}

func (mc *ContextBuilder) WithExternalServicesEndpointMap(endpointMap *EndpointMapBuilder) *ContextBuilder {
	mc.res.Mesh.ExternalServicesEndpointMap = endpointMap.Build()
	mc.res.ControlPlane.CLACache.(*xds.DummyCLACache).OutboundTargets = endpointMap.Build()
	return mc
}

func (mc *ContextBuilder) WithZone(zone string) *ContextBuilder {
	mc.res.ControlPlane.Zone = zone
	return mc
}

func (mc *ContextBuilder) WithResources(resources xds_context.Resources) *ContextBuilder {
	mc.res.Mesh.Resources = resources
	return mc
}

func (mc *ContextBuilder) WithMeshBuilder(mesh *builders.MeshBuilder) *ContextBuilder {
	mc.res.Mesh.Resource = mesh.Build()
	return mc
}

func (mc *ContextBuilder) WithMeshContext(mesh *xds_context.MeshContext) *ContextBuilder {
	mc.res.Mesh = *mesh
	return mc
}

func (mc *ContextBuilder) AddServiceProtocol(serviceName string, protocol core_mesh.Protocol) *ContextBuilder {
	if info, found := mc.res.Mesh.ServicesInformation[serviceName]; found {
		info.Protocol = protocol
	} else {
		mc.res.Mesh.ServicesInformation[serviceName] = &xds_context.ServiceInformation{
			Protocol: protocol,
		}
	}
	return mc
}

func (mc *ContextBuilder) AddExternalService(serviceName string) *ContextBuilder {
	if info, found := mc.res.Mesh.ServicesInformation[serviceName]; found {
		info.IsExternalService = true
	} else {
		mc.res.Mesh.ServicesInformation[serviceName] = &xds_context.ServiceInformation{
			IsExternalService: true,
		}
	}
	return mc
}
