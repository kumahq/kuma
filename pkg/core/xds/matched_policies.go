package xds

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

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

	// Dataplane -> Policy
	TrafficTrace *core_mesh.TrafficTraceResource
}
