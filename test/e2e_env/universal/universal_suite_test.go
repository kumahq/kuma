package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e_env/universal/api"
	"github.com/kumahq/kuma/test/e2e_env/universal/auth"
	"github.com/kumahq/kuma/test/e2e_env/universal/compatibility"
	"github.com/kumahq/kuma/test/e2e_env/universal/externalservices"
	"github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	"github.com/kumahq/kuma/test/e2e_env/universal/grpc"
	"github.com/kumahq/kuma/test/e2e_env/universal/healthcheck"
	"github.com/kumahq/kuma/test/e2e_env/universal/inspect"
	"github.com/kumahq/kuma/test/e2e_env/universal/intercp"
	"github.com/kumahq/kuma/test/e2e_env/universal/matching"
	"github.com/kumahq/kuma/test/e2e_env/universal/membership"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshaccesslog"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshexternalservice"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshfaultinjection"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshhealthcheck"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshloadbalancingstrategy"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshproxypatch"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshratelimit"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshretry"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshservice"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshtls"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshtrafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/universal/mtls"
	"github.com/kumahq/kuma/test/e2e_env/universal/observability"
	"github.com/kumahq/kuma/test/e2e_env/universal/projectedsatoken"
	"github.com/kumahq/kuma/test/e2e_env/universal/proxytemplate"
	"github.com/kumahq/kuma/test/e2e_env/universal/ratelimit"
	"github.com/kumahq/kuma/test/e2e_env/universal/reachableservices"
	"github.com/kumahq/kuma/test/e2e_env/universal/resilience"
	"github.com/kumahq/kuma/test/e2e_env/universal/retry"
	"github.com/kumahq/kuma/test/e2e_env/universal/timeout"
	"github.com/kumahq/kuma/test/e2e_env/universal/trafficlog"
	"github.com/kumahq/kuma/test/e2e_env/universal/trafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/universal/trafficroute"
	"github.com/kumahq/kuma/test/e2e_env/universal/transparentproxy"
	"github.com/kumahq/kuma/test/e2e_env/universal/virtualoutbound"
	"github.com/kumahq/kuma/test/e2e_env/universal/zoneegress"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Universal Suite")
}

var (
	_ = E2ESynchronizedBeforeSuite(universal.SetupAndGetState, universal.RestoreState)
	_ = SynchronizedAfterSuite(func() {}, universal.SynchronizedAfterSuite)
	_ = ReportAfterSuite("universal after suite", universal.AfterSuite)
)

var (
	_ = Describe("User Auth", auth.UserAuth)
	_ = Describe("DP Auth", auth.DpAuth, Ordered)
	_ = Describe("Offline Auth", auth.OfflineAuth, Ordered)
	_ = Describe("Gateway", gateway.Gateway, Ordered)
	_ = Describe("Gateway - Cross-mesh", gateway.CrossMeshGatewayOnUniversal, Ordered)
	_ = Describe("HealthCheck panic threshold", healthcheck.HealthCheckPanicThreshold, Ordered)
	_ = Describe("HealthCheck", healthcheck.Policy)
	_ = Describe("MeshHealthCheck panic threshold", meshhealthcheck.MeshHealthCheckPanicThreshold, Ordered)
	_ = Describe("MeshHealthCheck", meshhealthcheck.MeshHealthCheck)
	_ = Describe("Service Probes", healthcheck.ServiceProbes, Ordered)
	_ = Describe("External Services", externalservices.Policy, Ordered)
	_ = Describe("External Services through Zone Egress", externalservices.ThroughZoneEgress, Ordered)
	_ = Describe("Inspect", inspect.Inspect, Ordered)
	_ = Describe("Mesh External Services", meshexternalservice.MeshExternalService, Ordered)
	_ = Describe("MeshService", meshservice.MeshService, Ordered)
	_ = Describe("Applications Metrics", observability.ApplicationsMetrics, Ordered)
	_ = Describe("Tracing", observability.Tracing, Ordered)
	_ = Describe("MeshTrace", observability.PluginTest, Ordered)
	_ = Describe("Membership", membership.Membership, Ordered)
	_ = Describe("Traffic Logging", trafficlog.TCPLogging, Ordered)
	_ = Describe("MeshAccessLog", meshaccesslog.TestPlugin, Ordered)
	_ = Describe("Timeout", timeout.Policy, Ordered)
	_ = Describe("Retry", retry.Policy, Ordered)
	_ = Describe("MeshRetry", meshretry.HttpRetry, Ordered)
	_ = Describe("MeshRetry", meshretry.GrpcRetry, Ordered)
	_ = Describe("RateLimit", ratelimit.Policy, Ordered)
	_ = Describe("ProxyTemplate", proxytemplate.ProxyTemplate, Ordered)
	_ = Describe("MeshProxyPatch", meshproxypatch.MeshProxyPatch, Ordered)
	_ = Describe("Matching", matching.Matching, Ordered)
	_ = Describe("Mtls", mtls.Policy, Ordered)
	_ = Describe("Reachable Services", reachableservices.ReachableServices, Ordered)
	_ = Describe("Apis", api.Api, Ordered)
	_ = Describe("Traffic Permission", trafficpermission.TrafficPermission, Ordered)
	// TODO: fix the flaky test in the future https://github.com/kumahq/kuma/issues/11492
	_ = Describe("Traffic Route", trafficroute.TrafficRoute, Ordered, FlakeAttempts(3))
	_ = Describe("Zone Egress", zoneegress.ExternalServices, Ordered)
	_ = Describe("Virtual Outbound", virtualoutbound.VirtualOutbound, Ordered)
	_ = Describe("Transparent Proxy", transparentproxy.TransparentProxy, Ordered)
	_ = Describe("Mesh Traffic Permission", meshtrafficpermission.MeshTrafficPermissionUniversal, Ordered)
	_ = Describe("GRPC", grpc.GRPC, Ordered)
	_ = Describe("MeshRateLimit", meshratelimit.Policy, Ordered)
	_ = Describe("MeshTimeout", timeout.PluginTest, Ordered)
	_ = Describe("Projected Service Account Token", projectedsatoken.ProjectedServiceAccountToken, Ordered)
	_ = Describe("Compatibility", compatibility.UniversalCompatibility, Ordered)
	_ = Describe("Resilience", resilience.ResilienceUniversal, Ordered)
	_ = Describe("Leader Election", resilience.LeaderElectionPostgres, Ordered)
	_ = Describe("MeshFaultInjection", meshfaultinjection.Policy, Ordered)
	_ = Describe("MeshLoadBalancingStrategy", meshloadbalancingstrategy.Policy, Ordered)
	_ = Describe("InterCP Server", intercp.InterCP, Ordered)
	_ = Describe("Prometheus Metrics", observability.PrometheusMetrics, Ordered)
	_ = Describe("MeshTLS", meshtls.Policy, Ordered)
)
