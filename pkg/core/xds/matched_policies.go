package xds

import (
	"encoding"
	"fmt"
	"sort"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type PolicyItem interface {
	GetTargetRef() common_api.TargetRef
	GetDefault() interface{}
}

type PolicyItemWithMeta struct {
	PolicyItem
	core_model.ResourceMeta
}

type Policy interface {
	core_model.ResourceSpec
	GetTargetRef() common_api.TargetRef
}

type PolicyWithToList interface {
	Policy
	GetToList() []PolicyItem
}

type PolicyWithFromList interface {
	Policy
	GetFromList() []PolicyItem
}

type PolicyWithSingleItem interface {
	Policy
	GetPolicyItem() PolicyItem
}

type InboundListener struct {
	Address string
	Port    uint32
}

func BuildPolicyItemsWithMeta(items []PolicyItem, meta core_model.ResourceMeta) []PolicyItemWithMeta {
	var result []PolicyItemWithMeta
	for _, item := range items {
		result = append(result, PolicyItemWithMeta{
			PolicyItem:   item,
			ResourceMeta: meta,
		})
	}
	return result
}

// We need to implement TextMarshaler because InboundListener is used
// as a key for maps that are JSON encoded for logging.
var _ encoding.TextMarshaler = InboundListener{}

func (i InboundListener) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i InboundListener) String() string {
	return fmt.Sprintf("%s:%d", i.Address, i.Port)
}

type FromRules struct {
	Rules map[InboundListener]Rules
}

type ToRules struct {
	Rules Rules
}

type SingleItemRules struct {
	Rules Rules
}

// TypedMatchingPolicies all policies of this type matching
type TypedMatchingPolicies struct {
	Type              core_model.ResourceType
	InboundPolicies   map[mesh_proto.InboundInterface][]core_model.Resource
	OutboundPolicies  map[mesh_proto.OutboundInterface][]core_model.Resource
	ServicePolicies   map[ServiceName][]core_model.Resource
	DataplanePolicies []core_model.Resource
	FromRules         FromRules
	ToRules           ToRules
	SingleItemRules   SingleItemRules
}

type PluginOriginatedPolicies map[core_model.ResourceType]TypedMatchingPolicies

type MatchedPolicies struct {
	// Inbound(Listener) -> Policy
	TrafficPermissions TrafficPermissionMap
	FaultInjections    FaultInjectionMap
	RateLimitsInbound  InboundRateLimitsMap

	// Service(Cluster) -> Policy
	TrafficLogs     TrafficLogMap
	HealthChecks    HealthCheckMap
	CircuitBreakers CircuitBreakerMap
	Retries         RetryMap

	// Outbound(Listener) -> Policy
	Timeouts           TimeoutMap
	RateLimitsOutbound OutboundRateLimitsMap
	// Actual Envoy Configuration is generated without taking this TrafficRoutes into account
	TrafficRoutes RouteMap

	// Dataplane -> Policy
	TrafficTrace *core_mesh.TrafficTraceResource
	// Actual Envoy Configuration is generated without taking this ProxyTemplate into account
	ProxyTemplate *core_mesh.ProxyTemplateResource

	Dynamic PluginOriginatedPolicies
}

func (m *MatchedPolicies) OrderedDynamicPolicies() []core_model.ResourceType {
	var all []core_model.ResourceType
	for k := range m.Dynamic {
		all = append(all, k)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i] < all[j]
	})
	return all
}
