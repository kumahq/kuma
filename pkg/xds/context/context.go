package context

import (
	"encoding/base64"
	"fmt"
	"time"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
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
	LabelsToValuesToResourceIdentifier map[LabelValue]map[core_model.ResourceIdentifier]bool
)

// MeshContext contains shared data within one mesh that is required for generating XDS config.
// This data is the same for all data plane proxies within one mesh.
// If there is an information that can be precomputed and shared between all data plane proxies
// it should be put here. This way we can save CPU cycles of computing the same information.
type MeshContext struct {
	Hash                                string
	Resource                            *core_mesh.MeshResource
	Resources                           Resources
	DataplanesByName                    map[string]*core_mesh.DataplaneResource
	MeshServiceByIdentifier             map[core_model.ResourceIdentifier]*meshservice_api.MeshServiceResource
	MeshServicesByLabelByValue          LabelsToValuesToResourceIdentifier
	MeshExternalServiceByIdentifier     map[core_model.ResourceIdentifier]*meshexternalservice_api.MeshExternalServiceResource
	MeshExternalServicesByLabelByValue  LabelsToValuesToResourceIdentifier
	MeshMultiZoneServiceByIdentifier    map[core_model.ResourceIdentifier]*meshmzservice_api.MeshMultiZoneServiceResource
	MeshMultiZoneServicesByLabelByValue LabelsToValuesToResourceIdentifier
	EndpointMap                         xds.EndpointMap
	IngressEndpointMap                  xds.EndpointMap
	ExternalServicesEndpointMap         xds.EndpointMap
	CrossMeshEndpoints                  map[xds.MeshName]xds.EndpointMap
	VIPDomains                          []xds_types.VIPDomains
	VIPOutbounds                        xds_types.Outbounds
	ServicesInformation                 map[string]*ServiceInformation
	DataSourceLoader                    datasource.Loader
	ReachableServicesGraph              ReachableServicesGraph
}

type ServiceInformation struct {
	TLSReadiness      bool
	Protocol          core_mesh.Protocol
	IsExternalService bool
}

type ReachableBackends map[core_model.TypedResourceIdentifier]bool

// ResolveResourceIdentifier resolves one resource identifier based on the labels.
// If multiple resources match the labels, the oldest one is returned.
// The reason is that picking the oldest one is the less likely to break existing traffic after introducing new resources.
func (mc *MeshContext) ResolveResourceIdentifier(resType core_model.ResourceType, labels map[string]string) *core_model.ResourceIdentifier {
	if len(labels) == 0 {
		return nil
	}
	var oldestCreationTime *time.Time
	var oldestTri *core_model.TypedResourceIdentifier
	for _, tri := range mc.resolveResourceIdentifiersForLabels(string(resType), labels) {
		var resource core_model.Resource
		var found bool
		switch tri.ResourceType {
		case meshexternalservice_api.MeshExternalServiceType:
			resource, found = mc.MeshExternalServiceByIdentifier[tri.ResourceIdentifier]
		case meshservice_api.MeshServiceType:
			resource, found = mc.MeshServiceByIdentifier[tri.ResourceIdentifier]
		case meshmzservice_api.MeshMultiZoneServiceType:
			resource, found = mc.MeshMultiZoneServiceByIdentifier[tri.ResourceIdentifier]
		}
		if found {
			resCreationTime := resource.GetMeta().GetCreationTime()
			if oldestCreationTime == nil || resCreationTime.Before(*oldestCreationTime) {
				oldestCreationTime = &resCreationTime
				oldestTri = &tri
			}
		}
	}
	if oldestTri != nil {
		return &oldestTri.ResourceIdentifier
	}
	return nil
}

func (mc *MeshContext) GetReachableBackends(dataplane *core_mesh.DataplaneResource) *ReachableBackends {
	if dataplane.Spec.Networking.TransparentProxying.GetReachableBackends() == nil {
		if mc.Resource.Spec.MeshServicesEnabled() == mesh_proto.Mesh_MeshServices_ReachableBackends {
			return &ReachableBackends{}
		}
		return nil
	}
	reachableBackends := ReachableBackends{}
	for _, reachableBackend := range dataplane.Spec.Networking.TransparentProxying.GetReachableBackends().GetRefs() {
		if len(reachableBackend.Labels) > 0 {
			for _, tri := range mc.resolveResourceIdentifiersForLabels(reachableBackend.Kind, reachableBackend.Labels) {
				reachableBackends[tri] = true
			}
		} else {
			key := core_model.TypedResourceIdentifier{
				ResourceType: core_model.ResourceType(reachableBackend.Kind),
				ResourceIdentifier: core_model.TargetRefToResourceIdentifier(dataplane.GetMeta(), common_api.TargetRef{
					Name:      reachableBackend.Name,
					Namespace: reachableBackend.Namespace,
				}),
			}
			if port := reachableBackend.Port; port != nil {
				key.SectionName = fmt.Sprintf("%d", port.GetValue())
			}
			reachableBackends[key] = true
		}
	}
	return &reachableBackends
}

func (mc *MeshContext) resolveResourceIdentifiersForLabels(kind string, labels map[string]string) []core_model.TypedResourceIdentifier {
	var result []core_model.TypedResourceIdentifier
	reachable := mc.getResourceNamesForLabels(kind, labels)
	for ri, count := range reachable {
		tri := core_model.TypedResourceIdentifier{
			ResourceType:       core_model.ResourceType(kind),
			ResourceIdentifier: ri,
		}
		if count == len(labels) {
			result = append(result, tri)
		}
	}
	return result
}

func (mc *MeshContext) getResourceNamesForLabels(kind string, labels map[string]string) map[core_model.ResourceIdentifier]int {
	reachable := map[core_model.ResourceIdentifier]int{}
	for label, value := range labels {
		key := LabelValue{
			Label: label,
			Value: value,
		}
		var matchedResourceIdentifiers map[core_model.ResourceIdentifier]bool
		var found bool
		switch kind {
		case string(meshexternalservice_api.MeshExternalServiceType):
			matchedResourceIdentifiers, found = mc.MeshExternalServicesByLabelByValue[key]
		case string(meshservice_api.MeshServiceType):
			matchedResourceIdentifiers, found = mc.MeshServicesByLabelByValue[key]
		case string(meshmzservice_api.MeshMultiZoneServiceType):
			matchedResourceIdentifiers, found = mc.MeshMultiZoneServicesByLabelByValue[key]
		}
		if found {
			for ri := range matchedResourceIdentifiers {
				reachable[ri]++
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
