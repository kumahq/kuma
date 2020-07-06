package xds

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

// StreamID represents a stream opened by XDS
type StreamID = int64

type ProxyId struct {
	Mesh string
	Name string
}

func (id ProxyId) String() string {
	return fmt.Sprintf("%s.%s", id.Mesh, id.Name)
}

// ServiceName is a convenience type alias to clarify the meaning of string value.
type ServiceName = string

// RouteMap holds the most specific TrafficRoute for each outbound interface of a Dataplane.
type RouteMap map[mesh_proto.OutboundInterface]*mesh_core.TrafficRouteResource

// TagSelectorSet is a set of unique TagSelectors.
type TagSelectorSet []mesh_proto.TagSelector

// DestinationMap holds a set of selectors for all reachable Dataplanes grouped by service name.
// DestinationMap is based on ServiceName and not on the OutboundInterface because TrafficRoute can introduce new service destinations that were not included in a outbound section.
// Policies that match on outbound connections also match by service destination name and not outbound interface for the same reason.
type DestinationMap map[ServiceName]TagSelectorSet

// Endpoint holds routing-related information about a single endpoint.
type Endpoint struct {
	Target string
	Port   uint32
	Tags   map[string]string
	Weight uint32
}

// EndpointList is a list of Endpoints with convenience methods.
type EndpointList []Endpoint

// EndpointMap holds routing-related information about a set of endpoints grouped by service name.
type EndpointMap map[ServiceName][]Endpoint

// LogMap holds the most specific TrafficLog for each outbound interface of a Dataplane.
type LogMap map[ServiceName]*mesh_proto.LoggingBackend

// HealthCheckMap holds the most specific HealthCheck for each reachable service.
type HealthCheckMap map[ServiceName]*mesh_core.HealthCheckResource

// CircuitBreakerMap holds the most specific CircuitBreaker for each reachable service.
type CircuitBreakerMap map[ServiceName]*mesh_core.CircuitBreakerResource

// FaultInjectionMap holds the most specific FaultInjectionResource for each InboundInterface
type FaultInjectionMap map[mesh_proto.InboundInterface]*mesh_proto.FaultInjection

// TrafficPermissionMap holds the most specific TrafficPermissionResource for each InboundInterface
type TrafficPermissionMap map[mesh_proto.InboundInterface]*mesh_core.TrafficPermissionResource

type Proxy struct {
	Id                 ProxyId
	Dataplane          *mesh_core.DataplaneResource
	TrafficPermissions TrafficPermissionMap
	Logs               LogMap
	TrafficRoutes      RouteMap
	OutboundSelectors  DestinationMap
	OutboundTargets    EndpointMap
	HealthChecks       HealthCheckMap
	CircuitBreakers    CircuitBreakerMap
	TrafficTrace       *mesh_core.TrafficTraceResource
	TracingBackend     *mesh_proto.TracingBackend
	Metadata           *DataplaneMetadata
	FaultInjections    FaultInjectionMap
}

func (s TagSelectorSet) Add(new mesh_proto.TagSelector) TagSelectorSet {
	for _, old := range s {
		if new.Equal(old) {
			return s
		}
	}
	return append(s, new)
}

func (s TagSelectorSet) Matches(tags map[string]string) bool {
	for _, selector := range s {
		if selector.Matches(tags) {
			return true
		}
	}
	return false
}

func (l EndpointList) Filter(selector mesh_proto.TagSelector) EndpointList {
	var endpoints EndpointList
	for _, endpoint := range l {
		if selector.Matches(endpoint.Tags) {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

func BuildProxyId(mesh, name string, more ...string) (*ProxyId, error) {
	id := strings.Join(append([]string{mesh, name}, more...), ".")
	return ParseProxyIdFromString(id)
}

func ParseProxyId(node *envoy_core.Node) (*ProxyId, error) {
	if node == nil {
		return nil, errors.Errorf("Envoy node must not be nil")
	}
	return ParseProxyIdFromString(node.Id)
}

func ParseProxyIdFromString(id string) (*ProxyId, error) {
	parts := strings.SplitN(id, ".", 2)
	mesh := parts[0]
	if mesh == "" {
		return nil, errors.New("mesh must not be empty")
	}
	if len(parts) < 2 {
		return nil, errors.New("the name should be provided after the dot")
	}
	name := parts[1]
	if name == "" {
		return nil, errors.New("name must not be empty")
	}
	return &ProxyId{
		Mesh: mesh,
		Name: name,
	}, nil
}

func (id *ProxyId) ToResourceKey() core_model.ResourceKey {
	return core_model.ResourceKey{
		Name: id.Name,
		Mesh: id.Mesh,
	}
}

func FromResourceKey(key core_model.ResourceKey) ProxyId {
	return ProxyId{
		Mesh: key.Mesh,
		Name: key.Name,
	}
}
