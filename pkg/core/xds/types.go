package xds

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	util_tls "github.com/kumahq/kuma/pkg/tls"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
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

type MeshName = string

// RouteMap holds the most specific TrafficRoute for each outbound interface of a Dataplane.
type RouteMap map[mesh_proto.OutboundInterface]*core_mesh.TrafficRouteResource

// TimeoutMap holds the most specific TimeoutResource for each OutboundInterface
type TimeoutMap map[mesh_proto.OutboundInterface]*core_mesh.TimeoutResource

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
	ServerName         string
}

type Locality struct {
	Zone     string
	Priority uint32
}

// Endpoint holds routing-related information about a single endpoint.
type Endpoint struct {
	Target          string
	UnixDomainPath  string
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

// TrafficLogMap holds the most specific TrafficLog for each outbound interface of a Dataplane.
type TrafficLogMap map[ServiceName]*core_mesh.TrafficLogResource

// HealthCheckMap holds the most specific HealthCheck for each reachable service.
type HealthCheckMap map[ServiceName]*core_mesh.HealthCheckResource

// CircuitBreakerMap holds the most specific CircuitBreaker for each reachable service.
type CircuitBreakerMap map[ServiceName]*core_mesh.CircuitBreakerResource

// RetryMap holds the most specific Retry for each reachable service.
type RetryMap map[ServiceName]*core_mesh.RetryResource

// FaultInjectionMap holds all matched FaultInjectionResources for each InboundInterface
type FaultInjectionMap map[mesh_proto.InboundInterface][]*core_mesh.FaultInjectionResource

// TrafficPermissionMap holds the most specific TrafficPermissionResource for each InboundInterface
type TrafficPermissionMap map[mesh_proto.InboundInterface]*core_mesh.TrafficPermissionResource

// InboundRateLimitsMap holds all RateLimitResources for each InboundInterface
type InboundRateLimitsMap map[mesh_proto.InboundInterface][]*core_mesh.RateLimitResource

// OutboundRateLimitsMap holds the RateLimitResource for each OutboundInterface
type OutboundRateLimitsMap map[mesh_proto.OutboundInterface]*core_mesh.RateLimitResource

type RateLimitsMap struct {
	Inbound  InboundRateLimitsMap
	Outbound OutboundRateLimitsMap
}

type ExternalServicePermissionMap map[ServiceName]*core_mesh.TrafficPermissionResource

type ExternalServiceFaultInjectionMap map[ServiceName][]*core_mesh.FaultInjectionResource

type ExternalServiceRateLimitMap map[ServiceName][]*core_mesh.RateLimitResource

type CLACache interface {
	GetCLA(ctx context.Context, meshName, meshHash string, cluster envoy_common.Cluster, apiVersion envoy_common.APIVersion, endpointMap EndpointMap) (proto.Message, error)
}

// SocketAddressProtocol is the L4 protocol the listener should bind to
type SocketAddressProtocol int32

const (
	SocketAddressProtocolTCP SocketAddressProtocol = 0
	SocketAddressProtocolUDP SocketAddressProtocol = 1
)

// Proxy contains required data for generating XDS config that is specific to a data plane proxy.
// The data that is specific for the whole mesh should go into MeshContext.
type Proxy struct {
	Id                  ProxyId
	APIVersion          envoy_common.APIVersion // todo(jakubdyszkiewicz) consider moving APIVersion here. pkg/core should not depend on pkg/xds. It should be other way around.
	Dataplane           *core_mesh.DataplaneResource
	ZoneIngress         *core_mesh.ZoneIngressResource
	Metadata            *DataplaneMetadata
	Routing             Routing
	Policies            MatchedPolicies
	EnvoyAdminMTLSCerts ServerSideMTLSCerts

	// SecretsTracker allows us to track when a generator references a secret so
	// we can be sure to include only those secrets later on.
	SecretsTracker SecretsTracker

	// ZoneEgressProxy is available only when XDS is generated for ZoneEgress data plane proxy.
	ZoneEgressProxy *ZoneEgressProxy
	// ZoneIngressProxy is available only when XDS is generated for ZoneIngress data plane proxy.
	ZoneIngressProxy *ZoneIngressProxy
}

type ServerSideMTLSCerts struct {
	CaPEM      []byte
	ServerPair util_tls.KeyPair
}

type identityCertRequest struct {
	meshName string
}

func (r identityCertRequest) Name() string {
	return names.GetSecretName(xds_tls.IdentityCertResource, "secret", r.meshName)
}

type IdentityCertRequest interface {
	Name() string
}

type CaRequest interface {
	MeshName() []string
	Name() string
}

type caRequest struct {
	meshName string
}

type allInOneCaRequest struct {
	meshNames []string
}

func (r caRequest) Name() string {
	return names.GetSecretName(xds_tls.MeshCaResource, "secret", r.meshName)
}

func (r caRequest) MeshName() []string {
	return []string{r.meshName}
}

func (r allInOneCaRequest) Name() string {
	return names.GetSecretName(xds_tls.MeshCaResource, "secret", "all")
}

func (r allInOneCaRequest) MeshName() []string {
	return r.meshNames
}

// SecretsTracker provides a way to ask for a secret and keeps track of which are
// used, so that they can later be generated and included in the resources.
type SecretsTracker interface {
	RequestIdentityCert() IdentityCertRequest
	RequestCa(mesh string) CaRequest
	RequestAllInOneCa() CaRequest

	UsedIdentity() bool
	UsedCas() map[string]struct{}
	UsedAllInOne() bool
}

type secretsTracker struct {
	ownMesh   string
	allMeshes []string

	identity bool
	meshes   map[string]struct{}
	allInOne bool
}

func NewSecretsTracker(ownMesh string, allMeshes []string) SecretsTracker {
	return &secretsTracker{
		ownMesh:   ownMesh,
		allMeshes: allMeshes,

		meshes: map[string]struct{}{},
	}
}

func (st *secretsTracker) RequestIdentityCert() IdentityCertRequest {
	st.identity = true
	return &identityCertRequest{
		meshName: st.ownMesh,
	}
}

func (st *secretsTracker) RequestCa(mesh string) CaRequest {
	st.meshes[mesh] = struct{}{}
	return &caRequest{
		meshName: mesh,
	}
}

func (st *secretsTracker) RequestAllInOneCa() CaRequest {
	st.allInOne = true
	return &allInOneCaRequest{
		meshNames: st.allMeshes,
	}
}
func (st *secretsTracker) UsedIdentity() bool {
	return st.identity
}
func (st *secretsTracker) UsedCas() map[string]struct{} {
	return st.meshes
}
func (st *secretsTracker) UsedAllInOne() bool {
	return st.allInOne
}

type MeshResources struct {
	Mesh                           *core_mesh.MeshResource
	TrafficRoutes                  []*core_mesh.TrafficRouteResource
	ExternalServices               []*core_mesh.ExternalServiceResource
	ExternalServicePermissionMap   ExternalServicePermissionMap
	EndpointMap                    EndpointMap
	ExternalServiceFaultInjections ExternalServiceFaultInjectionMap
	ExternalServiceRateLimits      ExternalServiceRateLimitMap
}

type ZoneEgressProxy struct {
	ZoneEgressResource *core_mesh.ZoneEgressResource
	ZoneIngresses      []*core_mesh.ZoneIngressResource
	MeshResourcesList  []*MeshResources
}

type ZoneIngressProxy struct {
	TrafficRouteList *core_mesh.TrafficRouteResourceList
	GatewayRoutes    *core_mesh.MeshGatewayRouteResourceList
	MeshGateways     *core_mesh.MeshGatewayResourceList
}

type VIPDomains struct {
	Address string
	Domains []string
}

type Routing struct {
	TrafficRoutes   RouteMap
	OutboundTargets EndpointMap
	// ExternalServiceOutboundTargets contains endpoint map for direct access of external services (without egress)
	// Since we take into account TrafficPermission to exclude external services from the map,
	// it is specific for each data plane proxy.
	ExternalServiceOutboundTargets EndpointMap
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
	return e.Locality.Zone
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

// if false endpoint should be accessed through zoneIngress of other zone
func (e Endpoint) IsReachableFromZone(localZone string) bool {
	return e.Locality == nil || e.Locality.Zone == "" || e.Locality.Zone == localZone
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
