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
	RateLimitWsDefinition,
	TrafficPermissionWsDefinition,
	TrafficLogWsDefinition,
	TrafficRouteWsDefinition,
	TrafficTraceWsDefinition,
	FaultInjectionWsDefinition,
	CircuitBreakerWsDefinition,
	ZoneWsDefinition,
	ZoneInsightWsDefinition,
	SecretWsDefinition,
	GlobalSecretWsDefinition,
	RetryWsDefinition,
	TimeoutWsDefinition,
}
