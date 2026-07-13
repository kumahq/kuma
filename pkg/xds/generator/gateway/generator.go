package gateway

import (
	"fmt"
	"maps"

	envoy_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/match"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/route"
)

// ConnectionPolicyTypes specifies the resource types the gateway will
// bind for connection policies.
var ConnectionPolicyTypes = []model.ResourceType{
	core_mesh.CircuitBreakerType,
	core_mesh.FaultInjectionType,
	core_mesh.HealthCheckType,
	core_mesh.RateLimitType,
	core_mesh.RetryType,
	core_mesh.TimeoutType,
}

type GatewayHostInfo struct {
	Host GatewayHost
	// These are entries created internally in this plugin by MeshGatewayRoute
	// before the Mesh*Route policies run
	meshGatewayRouteEntries []route.Entry
	// This are entries created by new Mesh*Route policies
	routeEntries []route.Entry
}

func (i GatewayHostInfo) Entries() []route.Entry {
	// We need to return one or the other because the gateway plugin doesn't
	// know about Mesh*Routes and generates a 404 entry.
	if len(i.routeEntries) > 0 {
		return i.routeEntries
	}
	return i.meshGatewayRouteEntries
}

func (i *GatewayHostInfo) AppendEntries(entries []route.Entry) {
	i.routeEntries = append(i.routeEntries, entries...)
}

type GatewayHost struct {
	Hostname string
	Routes   []*core_mesh.MeshGatewayRouteResource
	Policies map[model.ResourceType][]match.RankedPolicy
	// Contains MeshGateway, Listener and Dataplane object tags
	Tags mesh_proto.TagSelector
}

type GatewayListenerHostname struct {
	Hostname  string
	Protocol  mesh_proto.MeshGateway_Listener_Protocol
	TLS       *mesh_proto.MeshGateway_TLS_Conf
	HostInfos []GatewayHostInfo
}

func (h GatewayListenerHostname) EnvoyRouteName(envoyListenerName string) string {
	switch h.Protocol {
	case mesh_proto.MeshGateway_Listener_TCP, mesh_proto.MeshGateway_Listener_HTTP:
		return envoyListenerName + ":*"
	default:
		return envoyListenerName + ":" + h.Hostname
	}
}

type GatewayListener struct {
	Port              uint32
	Protocol          mesh_proto.MeshGateway_Listener_Protocol
	EnvoyListenerName string
	// CrossMesh is important because for generation we need to treat such a
	// listener as if we have HTTPS with the Mesh cert for this Dataplane
	CrossMesh bool
	Resources *mesh_proto.MeshGateway_Listener_Resources // TODO verify these don't conflict when merging
}

// GatewayListenerInfo holds everything needed to generate resources for a
// listener.
type GatewayListenerInfo struct {
	Proxy             *core_xds.Proxy
	Gateway           *core_mesh.MeshGatewayResource
	ExternalServices  *core_mesh.ExternalServiceResourceList
	OutboundEndpoints core_xds.EndpointMap

	Listener          GatewayListener
	ListenerHostnames []GatewayListenerHostname
}

// FilterChainGenerator is responsible for handling the filter chain for
// a specific protocol.
// A FilterChainGenerator can be host-specific or shared amongst hosts.
type FilterChainGenerator interface {
	Generate(xds_context.Context, GatewayListenerInfo) (*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error)
}

// ExtractGatewayListeners returns the gateway listener info previously
// stored on the proxy by SetGatewayListeners.
func ExtractGatewayListeners(proxy *core_xds.Proxy) map[uint32]GatewayListenerInfo {
	ext := proxy.RuntimeExtensions[metadata.PluginName]
	if ext == nil {
		return nil
	}
	return ext.(map[uint32]GatewayListenerInfo)
}

// SetGatewayListeners assumes that exactly one plugin has authority over a
// single port.
func SetGatewayListeners(proxy *core_xds.Proxy, listenerInfoPerPort map[uint32]GatewayListenerInfo) {
	existingListeners := map[uint32]GatewayListenerInfo{}
	if ext := proxy.RuntimeExtensions[metadata.PluginName]; ext != nil {
		existingListeners = ext.(map[uint32]GatewayListenerInfo)
	}
	maps.Copy(existingListeners, listenerInfoPerPort)
	proxy.RuntimeExtensions[metadata.PluginName] = existingListeners
}

func GenerateRTDS(limits []RuntimeResoureLimitListener) *core_xds.Resource {
	layer := map[string]any{}
	for _, limit := range limits {
		layer[fmt.Sprintf("envoy.resource_limits.listener.%s.connection_limit", limit.Name)] = limit.ConnectionLimit
	}

	res := &core_xds.Resource{
		Name:   "gateway.listeners",
		Origin: metadata.OriginGateway,
		Resource: &envoy_service_runtime_v3.Runtime{
			Name:  "gateway.listeners",
			Layer: util_proto.MustStruct(layer),
		},
	}

	return res
}
