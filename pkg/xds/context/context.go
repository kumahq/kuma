package context

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

type Context struct {
	ControlPlane *ControlPlaneContext
	Mesh         MeshContext
}

type ConnectionInfo struct {
	// Authority defines the URL that was used by the data plane to connect to the control plane
	Authority string
}

// ControlPlaneContext contains shared global data and components that are required for generating XDS
// This data is the same regardless of a data plane proxy and mesh we are generating the data for.
type ControlPlaneContext struct {
	CLACache envoy.CLACache
	Secrets  secrets.Secrets
	Zone     string
}

// MeshContext contains shared data within one mesh that is required for generating XDS config.
// This data is the same for all data plane proxies within one mesh.
// If there is an information that can be precomputed and shared between all data plane proxies
// it should be put here. This way we can save CPU cycles of computing the same information.
type MeshContext struct {
	Hash                string
	Resource            *core_mesh.MeshResource
	Resources           Resources
	DataplanesByName    map[string]*core_mesh.DataplaneResource
	EndpointMap         xds.EndpointMap
	CrossMeshEndpoints  map[xds.MeshName]xds.EndpointMap
	VIPDomains          []xds.VIPDomains
	VIPOutbounds        []*mesh_proto.Dataplane_Networking_Outbound
	ServiceTLSReadiness map[string]bool
	DataSourceLoader    datasource.Loader
}

func (mc *MeshContext) GetTracingBackend(tt *core_mesh.TrafficTraceResource) *mesh_proto.TracingBackend {
	if tt == nil {
		return nil
	}
	if tb := mc.Resource.GetTracingBackend(tt.Spec.GetConf().GetBackend()); tb == nil {
		core.Log.Info("Tracing backend is not found. Ignoring.",
			"backendName", tt.Spec.GetConf().GetBackend(),
			"trafficTraceName", tt.GetMeta().GetName(),
			"trafficTraceMesh", tt.GetMeta().GetMesh())
		return nil
	} else {
		return tb
	}
}

func (mc *MeshContext) GetLoggingBackend(tl *core_mesh.TrafficLogResource) *mesh_proto.LoggingBackend {
	if tl == nil {
		return nil
	}
	if lb := mc.Resource.GetLoggingBackend(tl.Spec.GetConf().GetBackend()); lb == nil {
		core.Log.Info("Logging backend is not found. Ignoring.",
			"backendName", tl.Spec.GetConf().GetBackend(),
			"trafficLogName", tl.GetMeta().GetName(),
			"trafficLogMesh", tl.GetMeta().GetMesh())
		return nil
	} else {
		return lb
	}
}

// MeshContexts is an aggregate of all MeshContexts across all meshes
type MeshContexts struct {
	Hash               string
	Meshes             []*core_mesh.MeshResource
	MeshContextsByName map[string]MeshContext
	ZoneEgressByName   map[string]*core_mesh.ZoneEgressResource
}

func (m MeshContexts) AllDataplanes() []*core_mesh.DataplaneResource {
	var resources []*core_mesh.DataplaneResource
	for _, meshCtx := range m.MeshContextsByName {
		resources = append(resources, meshCtx.Resources.Dataplanes().Items...)
	}
	return resources
}

func (m MeshContexts) ZoneEgresses() []*core_mesh.ZoneEgressResource {
	for _, meshCtx := range m.MeshContextsByName {
		return meshCtx.Resources.ZoneEgresses().Items // all mesh contexts has the same list
	}
	return nil
}

func (m MeshContexts) ZoneIngresses() []*core_mesh.ZoneIngressResource {
	for _, meshCtx := range m.MeshContextsByName {
		return meshCtx.Resources.ZoneIngresses().Items // all mesh contexts has the same list
	}
	return nil
}

func (m MeshContexts) AllMeshGateways() []*core_mesh.MeshGatewayResource {
	var resources []*core_mesh.MeshGatewayResource
	for _, meshCtx := range m.MeshContextsByName {
		resources = append(resources, meshCtx.Resources.MeshGateways().Items...)
	}
	return resources
}
