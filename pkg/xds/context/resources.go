package context

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type Resources map[core_model.ResourceType]core_model.ResourceList

func (r Resources) ListOrEmpty(resourceType core_model.ResourceType) core_model.ResourceList {
	list, ok := r[resourceType]
	if !ok {
		list, err := registry.Global().NewList(resourceType)
		if err != nil {
			panic(err)
		}
		return list
	}
	return list
}

func (r Resources) ExternalServices() *core_mesh.ExternalServiceResourceList {
	return r.ListOrEmpty(core_mesh.ExternalServiceType).(*core_mesh.ExternalServiceResourceList)
}

func (r Resources) HealthChecks() *core_mesh.HealthCheckResourceList {
	return r.ListOrEmpty(core_mesh.HealthCheckType).(*core_mesh.HealthCheckResourceList)
}

func (r Resources) TrafficTraces() *core_mesh.TrafficTraceResourceList {
	return r.ListOrEmpty(core_mesh.TrafficTraceType).(*core_mesh.TrafficTraceResourceList)
}

func (r Resources) TrafficRoutes() *core_mesh.TrafficRouteResourceList {
	return r.ListOrEmpty(core_mesh.TrafficRouteType).(*core_mesh.TrafficRouteResourceList)
}

func (r Resources) Retries() *core_mesh.RetryResourceList {
	return r.ListOrEmpty(core_mesh.RetryType).(*core_mesh.RetryResourceList)
}

func (r Resources) TrafficPermissions() *core_mesh.TrafficPermissionResourceList {
	return r.ListOrEmpty(core_mesh.TrafficPermissionType).(*core_mesh.TrafficPermissionResourceList)
}

func (r Resources) TrafficLogs() *core_mesh.TrafficLogResourceList {
	return r.ListOrEmpty(core_mesh.TrafficLogType).(*core_mesh.TrafficLogResourceList)
}

func (r Resources) FaultInjections() *core_mesh.FaultInjectionResourceList {
	return r.ListOrEmpty(core_mesh.FaultInjectionType).(*core_mesh.FaultInjectionResourceList)
}

func (r Resources) Timeouts() *core_mesh.TimeoutResourceList {
	return r.ListOrEmpty(core_mesh.TimeoutType).(*core_mesh.TimeoutResourceList)
}

func (r Resources) RateLimits() *core_mesh.RateLimitResourceList {
	return r.ListOrEmpty(core_mesh.RateLimitType).(*core_mesh.RateLimitResourceList)
}

func (r Resources) CircuitBreakers() *core_mesh.CircuitBreakerResourceList {
	return r.ListOrEmpty(core_mesh.CircuitBreakerType).(*core_mesh.CircuitBreakerResourceList)
}

func (r Resources) ServiceInsights() *core_mesh.ServiceInsightResourceList {
	return r.ListOrEmpty(core_mesh.ServiceInsightType).(*core_mesh.ServiceInsightResourceList)
}

func (r Resources) ZoneIngresses() *core_mesh.ZoneIngressResourceList {
	return r.ListOrEmpty(core_mesh.ZoneIngressType).(*core_mesh.ZoneIngressResourceList)
}

func (r Resources) ZoneEgresses() *core_mesh.ZoneEgressResourceList {
	return r.ListOrEmpty(core_mesh.ZoneEgressType).(*core_mesh.ZoneEgressResourceList)
}

func (r Resources) Dataplanes() *core_mesh.DataplaneResourceList {
	return r.ListOrEmpty(core_mesh.DataplaneType).(*core_mesh.DataplaneResourceList)
}

func (r Resources) Gateways() *core_mesh.MeshGatewayResourceList {
	return r.ListOrEmpty(core_mesh.MeshGatewayType).(*core_mesh.MeshGatewayResourceList)
}

func (r Resources) GatewayRoutes() *core_mesh.MeshGatewayRouteResourceList {
	return r.ListOrEmpty(core_mesh.MeshGatewayRouteType).(*core_mesh.MeshGatewayRouteResourceList)
}

func (r Resources) ProxyTemplates() *core_mesh.ProxyTemplateResourceList {
	return r.ListOrEmpty(core_mesh.ProxyTemplateType).(*core_mesh.ProxyTemplateResourceList)
}
