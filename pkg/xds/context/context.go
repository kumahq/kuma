package context

import (
	"encoding/base64"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/k8s"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

var logger = core.Log.WithName("xds").WithName("context")

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
	CLACache        envoy.CLACache
	Secrets         secrets.Secrets
	Zone            string
	SystemNamespace string
}

// GlobalContext holds resources that are Global
type GlobalContext struct {
	ResourceMap ResourceMap
	hash        []byte
}

// Hash base64 version of the hash mostly used for testing
func (g GlobalContext) Hash() string {
	return base64.StdEncoding.EncodeToString(g.hash)
}

// BaseMeshContext holds for a Mesh a set of resources that are changing less often (policies, external services...)
type BaseMeshContext struct {
	Mesh        *core_mesh.MeshResource
	ResourceMap ResourceMap
	hash        []byte
}

// Hash base64 version of the hash mostly useed for testing
func (g BaseMeshContext) Hash() string {
	return base64.StdEncoding.EncodeToString(g.hash)
}

type LabelValue struct {
	Label string
	Value string
}
type (
	ServiceName                  string
	LabelsToValuesToServiceNames map[LabelValue]map[ServiceName]bool
)

// MeshContext contains shared data within one mesh that is required for generating XDS config.
// This data is the same for all data plane proxies within one mesh.
// If there is an information that can be precomputed and shared between all data plane proxies
// it should be put here. This way we can save CPU cycles of computing the same information.
type MeshContext struct {
	Hash                                    string
	Resource                                *core_mesh.MeshResource
	Resources                               Resources
	DataplanesByName                        map[string]*core_mesh.DataplaneResource
	MeshServiceByName                       map[string]*meshservice_api.MeshServiceResource
	MeshServiceNamesByLabelByValue          LabelsToValuesToServiceNames
	MeshExternalServiceByName               map[string]*meshexternalservice_api.MeshExternalServiceResource
	MeshExternalServiceNamesByLabelByValue  LabelsToValuesToServiceNames
	MeshMultiZoneServiceByName              map[string]*meshmzservice_api.MeshMultiZoneServiceResource
	MeshMultiZoneServiceNamesByLabelByValue LabelsToValuesToServiceNames
	EndpointMap                             xds.EndpointMap
	IngressEndpointMap                      xds.EndpointMap
	ExternalServicesEndpointMap             xds.EndpointMap
	CrossMeshEndpoints                      map[xds.MeshName]xds.EndpointMap
	VIPDomains                              []xds.VIPDomains
	VIPOutbounds                            xds.Outbounds
	ServicesInformation                     map[string]*ServiceInformation
	DataSourceLoader                        datasource.Loader
	ReachableServicesGraph                  ReachableServicesGraph
}

type ServiceInformation struct {
	TLSReadiness      bool
	Protocol          core_mesh.Protocol
	IsExternalService bool
}

type BackendKey struct {
	Kind string
	Name string
	Port uint32
}

type ReachableBackends map[BackendKey]bool

func (mc *MeshContext) GetReachableBackends(dataplane *core_mesh.DataplaneResource) *ReachableBackends {
	if dataplane.Spec.Networking.TransparentProxying.GetReachableBackends() == nil {
		return nil
	}
	reachableBackends := ReachableBackends{}
	for _, reachableBackend := range dataplane.Spec.Networking.TransparentProxying.GetReachableBackends().GetRefs() {
		key := BackendKey{Kind: reachableBackend.Kind}
		name := ""
		if reachableBackend.Name != "" {
			name = reachableBackend.Name
		}
		if reachableBackend.Namespace != "" {
			name = k8s.K8sNamespacedNameToCoreName(name, reachableBackend.Namespace)
		}
		key.Name = name
		if reachableBackend.Port != nil {
			key.Port = reachableBackend.Port.GetValue()
		}
		if len(reachableBackend.Labels) > 0 {
			reachable := mc.getResourceNamesForLabels(reachableBackend.Kind, reachableBackend.Labels)
			for name, count := range reachable {
				if count == len(reachableBackend.Labels) {
					reachableBackends[BackendKey{
						Kind: reachableBackend.Kind,
						Name: name,
					}] = true
				}
			}
		}
		if name != "" {
			reachableBackends[key] = true
		}
	}
	return &reachableBackends
}

func (mc *MeshContext) getResourceNamesForLabels(kind string, labels map[string]string) map[string]int {
	reachable := map[string]int{}
	for label, value := range labels {
		key := LabelValue{
			Label: label,
			Value: value,
		}
		var matchedServiceNames map[ServiceName]bool
		var found bool
		switch kind {
		case string(meshexternalservice_api.MeshExternalServiceType):
			matchedServiceNames, found = mc.MeshExternalServiceNamesByLabelByValue[key]
		case string(meshservice_api.MeshServiceType):
			matchedServiceNames, found = mc.MeshServiceNamesByLabelByValue[key]
		case string(meshmzservice_api.MeshMultiZoneServiceType):
			matchedServiceNames, found = mc.MeshMultiZoneServiceNamesByLabelByValue[key]
		}
		if found {
			for serviceName := range matchedServiceNames {
				reachable[string(serviceName)]++
			}
		}
	}
	return reachable
}

func (mc *MeshContext) GetTracingBackend(tt *core_mesh.TrafficTraceResource) *mesh_proto.TracingBackend {
	if tt == nil {
		return nil
	}
	if tb := mc.Resource.GetTracingBackend(tt.Spec.GetConf().GetBackend()); tb == nil {
		logger.Info("Tracing backend is not found. Ignoring.",
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
		logger.Info("Logging backend is not found. Ignoring.",
			"backendName", tl.Spec.GetConf().GetBackend(),
			"trafficLogName", tl.GetMeta().GetName(),
			"trafficLogMesh", tl.GetMeta().GetMesh())
		return nil
	} else {
		return lb
	}
}

func (mc *MeshContext) GetServiceProtocol(serviceName string) core_mesh.Protocol {
	if info, found := mc.ServicesInformation[serviceName]; found {
		return info.Protocol
	}
	return core_mesh.ProtocolUnknown
}

func (mc *MeshContext) IsExternalService(serviceName string) bool {
	if info, found := mc.ServicesInformation[serviceName]; found {
		return info.IsExternalService
	}
	return false
}

func (mc *MeshContext) GetTLSReadiness() map[string]bool {
	tlsReady := map[string]bool{}
	for serviceName, info := range mc.ServicesInformation {
		if info != nil {
			tlsReady[serviceName] = info.TLSReadiness
		} else {
			tlsReady[serviceName] = false
		}
	}
	return tlsReady
}

// AggregatedMeshContexts is an aggregate of all MeshContext across all meshes
type AggregatedMeshContexts struct {
	Hash               string
	Meshes             []*core_mesh.MeshResource
	MeshContextsByName map[string]MeshContext
	ZoneEgressByName   map[string]*core_mesh.ZoneEgressResource
}

// MustGetMeshContext panics if there is no mesh context for given mesh. Call it when iterating over .Meshes
// There is a guarantee that for every Mesh in .Meshes there is a MeshContext.
func (m AggregatedMeshContexts) MustGetMeshContext(meshName string) MeshContext {
	meshCtx, ok := m.MeshContextsByName[meshName]
	if !ok {
		panic("there should be a corresponding mesh context for every mesh in mesh contexts")
	}
	return meshCtx
}

func (m AggregatedMeshContexts) AllDataplanes() []*core_mesh.DataplaneResource {
	var resources []*core_mesh.DataplaneResource
	for _, mesh := range m.Meshes {
		meshCtx := m.MustGetMeshContext(mesh.Meta.GetName())
		resources = append(resources, meshCtx.Resources.Dataplanes().Items...)
	}
	return resources
}

func (m AggregatedMeshContexts) ZoneEgresses() []*core_mesh.ZoneEgressResource {
	for _, meshCtx := range m.MeshContextsByName {
		return meshCtx.Resources.ZoneEgresses().Items // all mesh contexts has the same list
	}
	return nil
}

func (m AggregatedMeshContexts) ZoneIngresses() []*core_mesh.ZoneIngressResource {
	for _, meshCtx := range m.MeshContextsByName {
		return meshCtx.Resources.ZoneIngresses().Items // all mesh contexts has the same list
	}
	return nil
}

func (m AggregatedMeshContexts) AllMeshGateways() []*core_mesh.MeshGatewayResource {
	var resources []*core_mesh.MeshGatewayResource
	for _, mesh := range m.Meshes {
		meshCtx := m.MustGetMeshContext(mesh.Meta.GetName())
		resources = append(resources, meshCtx.Resources.MeshGateways().Items...)
	}
	return resources
}
