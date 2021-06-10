package definitions

var All = append(
	DefaultCRUDLEndpoints,
	ServiceInsightWsDefinition,
)

var DefaultCRUDLEndpoints = []ResourceWsDefinition{
	MeshWsDefinition,
	MeshInsightWsDefinition,
	DataplaneWsDefinition,
	DataplaneInsightWsDefinition,
	ExternalServiceWsDefinition,
	HealthCheckWsDefinition,
	ProxyTemplateWsDefinition,
	TrafficPermissionWsDefinition,
	TrafficLogWsDefinition,
	TrafficRouteWsDefinition,
	TrafficTraceWsDefinition,
	FaultInjectionWsDefinition,
	CircuitBreakerWsDefinition,
	ZoneWsDefinition,
	ZoneInsightWsDefinition,
	ZoneIngressWsDefinition,
	SecretWsDefinition,
	GlobalSecretWsDefinition,
	RetryWsDefinition,
	TimeoutWsDefinition,
}
