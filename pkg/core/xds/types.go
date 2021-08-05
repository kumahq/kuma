package xds

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

// StreamID represents a stream opened by XDS
type StreamID = int64

type ProxyId struct {
	mesh string
	name string
}

func (id *ProxyId) String() string {
	return fmt.Sprintf("%s.%s", id.mesh, id.name)
}

func (id *ProxyId) ToResourceKey() core_model.ResourceKey {
	return core_model.ResourceKey{
		Name: id.name,
		Mesh: id.mesh,
	}
}

// ServiceName is a convenience type alias to clarify the meaning of string value.
type ServiceName = string

// RouteMap holds the most specific TrafficRoute for each outbound interface of a Dataplane.
type RouteMap map[mesh_proto.OutboundInterface]*mesh_core.TrafficRouteResource

// TimeoutMap holds the most specific TimeoutResource for each OutboundInterface
type TimeoutMap map[mesh_proto.OutboundInterface]*mesh_core.TimeoutResource

// TagSelectorSet is a set of unique TagSelectors.
type TagSelectorSet []mesh_proto.TagSelector

// DestinationMap holds a set of selectors for all reachable Dataplanes grouped by service name.
// DestinationMap is based on ServiceName and not on the OutboundInterface because TrafficRoute can introduce new service destinations that were not included in a outbound section.
// Policies that match on outbound connections also match by service destination name and not outbound interface for the same reason.
type DestinationMap map[ServiceName]TagSelectorSet

type ExternalService struct {
	TLSEnabled         bool
	CaCert             []byte
	ClientCert         []byte
	ClientKey          []byte
	AllowRenegotiation bool
}

type Locality struct {
	Region   string
	Zone     string
	SubZone  string
	Priority uint32
}

// Endpoint holds routing-related information about a single endpoint.
type Endpoint struct {
	Target          string
	Port            uint32
	Tags            map[string]string
	Weight          uint32
	Locality        *Locality
	ExternalService *ExternalService
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

// RetryMap holds the most specific Retry for each reachable service.
type RetryMap map[ServiceName]*mesh_core.RetryResource

// FaultInjectionMap holds the most specific FaultInjectionResource for each InboundInterface
type FaultInjectionMap map[mesh_proto.InboundInterface]*mesh_proto.FaultInjection

// TrafficPermissionMap holds the most specific TrafficPermissionResource for each InboundInterface
type TrafficPermissionMap map[mesh_proto.InboundInterface]*mesh_core.TrafficPermissionResource

// RateLimitsMap holds all RateLimitResources for each InboundInterface
type RateLimitsMap map[mesh_proto.InboundInterface][]*mesh_proto.RateLimit

type CLACache interface {
	GetCLA(ctx context.Context, meshName, meshHash string, cluster envoy_common.Cluster, apiVersion envoy_common.APIVersion) (proto.Message, error)
}

// SocketAddressProtocol is the L4 protocol the listener should bind to
type SocketAddressProtocol int32

const (
	SocketAddressProtocolTCP SocketAddressProtocol = 0
	SocketAddressProtocolUDP SocketAddressProtocol = 1
)

type Proxy struct {
	Id          ProxyId
	APIVersion  envoy_common.APIVersion // todo(jakubdyszkiewicz) consider moving APIVersion here. pkg/core should not depend on pkg/xds. It should be other way around.
	Dataplane   *mesh_core.DataplaneResource
	ZoneIngress *mesh_core.ZoneIngressResource
	Metadata    *DataplaneMetadata
	Routing     Routing
	Policies    MatchedPolicies
}

type VIPDomains struct {
	Address string
	Domains []string
}

type Routing struct {
	TrafficRoutes   RouteMap
	OutboundTargets EndpointMap
	VipDomains      []VIPDomains

	// todo(lobkovilya): split Proxy struct into DataplaneProxy and IngressProxy
	// TrafficRouteList is used only for generating configs for Ingress.
	TrafficRouteList *mesh_core.TrafficRouteResourceList
}

type MatchedPolicies struct {
	TrafficPermissions TrafficPermissionMap
	Logs               LogMap
	HealthChecks       HealthCheckMap
	CircuitBreakers    CircuitBreakerMap
	Retries            RetryMap
	TrafficTrace       *mesh_core.TrafficTraceResource
	TracingBackend     *mesh_proto.TracingBackend
	FaultInjections    FaultInjectionMap
	Timeouts           TimeoutMap
	RateLimits         RateLimitsMap
}

type CaSecret struct {
	PemCerts [][]byte
}

type IdentitySecret struct {
	PemCerts [][]byte
	PemKey   []byte
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

func (e Endpoint) IsExternalService() bool {
	return e.ExternalService != nil
}

func (e Endpoint) LocalityString() string {
	if e.Locality == nil {
		return ""
	}
	return e.Locality.Region + "/" + e.Locality.Zone + "/" + e.Locality.SubZone
}

func (e Endpoint) HasLocality() bool {
	return e.Locality != nil
}

// ContainsTags returns 'true' if for every key presented both in 'tags' and 'Endpoint#Tags'
// values are equal
func (e Endpoint) ContainsTags(tags map[string]string) bool {
	for otherKey, otherValue := range tags {
		endpointValue, ok := e.Tags[otherKey]
		if !ok || otherValue != endpointValue {
			return false
		}
	}
	return true
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

func BuildProxyId(mesh, name string) *ProxyId {
	return &ProxyId{
		name: name,
		mesh: mesh,
	}
}

func ParseProxyIdFromString(id string) (*ProxyId, error) {
	if id == "" {
		return nil, errors.Errorf("Envoy ID must not be nil")
	}
	parts := strings.SplitN(id, ".", 2)
	mesh := parts[0]
	// when proxy is an ingress mesh is empty
	if len(parts) < 2 {
		return nil, errors.New("the name should be provided after the dot")
	}
	name := parts[1]
	if name == "" {
		return nil, errors.New("name must not be empty")
	}
	return &ProxyId{
		mesh: mesh,
		name: name,
	}, nil
}

func FromResourceKey(key core_model.ResourceKey) ProxyId {
	return ProxyId{
		mesh: key.Mesh,
		name: key.Name,
	}
}
