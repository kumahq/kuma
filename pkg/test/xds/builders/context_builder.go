package builders

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/test/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type ContextBuilder struct {
	res *xds_context.Context
}

func Context() *ContextBuilder {
	return &ContextBuilder{
		res: &xds_context.Context{
			Mesh: xds_context.MeshContext{
				Resource:            samples.MeshDefault(),
				EndpointMap:         map[core_xds.ServiceName][]core_xds.Endpoint{},
				ServicesInformation: map[string]*xds_context.ServiceInformation{},
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

func (mc *ContextBuilder) WithZone(zone string) *ContextBuilder {
	mc.res.ControlPlane.Zone = zone
	return mc
}

func (mc *ContextBuilder) WithResources(resources xds_context.Resources) *ContextBuilder {
	mc.res.Mesh.Resources = resources
	return mc
}

func (mc *ContextBuilder) WithMesh(mesh *builders.MeshBuilder) *ContextBuilder {
	mc.res.Mesh.Resource = mesh.Build()
	return mc
}

func (mc *ContextBuilder) AddServiceProtocol(serviceName string, protocol core_mesh.Protocol) *ContextBuilder {
	if info, found := mc.res.Mesh.ServicesInformation[serviceName]; found {
		info.Protocol = protocol
		mc.res.Mesh.ServicesInformation[serviceName] = info
	} else {
		mc.res.Mesh.ServicesInformation[serviceName] = &xds_context.ServiceInformation{
			Protocol: protocol,
		}
	}
	return mc
}
