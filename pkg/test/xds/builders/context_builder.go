package builders

import (
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmultizoneservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
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
				BaseMeshContext:     &xds_context.BaseMeshContext{},
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
	resMap := xds_context.ResourceMap{
		meshservice_api.MeshServiceType:                   mc.res.Mesh.Resources.MeshServices(),
		meshexternalservice_api.MeshExternalServiceType:   mc.res.Mesh.Resources.MeshExternalServices(),
		meshmultizoneservice_api.MeshMultiZoneServiceType: mc.res.Mesh.Resources.MeshMultiZoneServices(),
	}
	mc.res.Mesh.BaseMeshContext.DestinationIndex = xds_context.NewDestinationIndex(resMap)
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

func (mc *ContextBuilder) WithMeshLocalResources(rs []model.Resource) *ContextBuilder {
	mc.res.Mesh.Resources = xds_context.Resources{MeshLocalResources: map[model.ResourceType]model.ResourceList{}}
	for _, p := range rs {
		if _, ok := mc.res.Mesh.Resources.MeshLocalResources[p.Descriptor().Name]; !ok {
			mc.res.Mesh.Resources.MeshLocalResources[p.Descriptor().Name] = registry.Global().MustNewList(p.Descriptor().Name)
		}
		Expect(mc.res.Mesh.Resources.MeshLocalResources[p.Descriptor().Name].AddItem(p)).To(Succeed())
	}
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
