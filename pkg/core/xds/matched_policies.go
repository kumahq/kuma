package xds

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
)

// TypedMatchingPolicies all policies of this type matching
type TypedMatchingPolicies struct {
	Type              core_model.ResourceType
	InboundPolicies   map[mesh_proto.InboundInterface][]core_model.Resource
	OutboundPolicies  map[mesh_proto.OutboundInterface][]core_model.Resource
	ServicePolicies   map[ServiceName][]core_model.Resource
	DataplanePolicies []core_model.Resource
	FromRules         rules.FromRules
	ToRules           rules.ToRules
	GatewayRules      rules.GatewayRules
	SingleItemRules   rules.SingleItemRules
	Warnings          []string
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
