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
	"github.com/kumahq/kuma/test/e2e_env/universal/matching"
	"github.com/kumahq/kuma/test/e2e_env/universal/membership"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshaccesslog"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshfaultinjection"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshhealthcheck"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshproxypatch"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshratelimit"
	"github.com/kumahq/kuma/test/e2e_env/universal/meshretry"
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
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Universal Suite")
}

var (
	_ = SynchronizedBeforeSuite(universal.SetupAndGetState, universal.RestoreState)
	_ = BeforeEach(universal.RememberSpecID)
	_ = AfterSuite(universal.WriteLogsIfFailed)
	_ = AfterEach(universal.WriteLogsIfFailed)
)

var (
	_ = PDescribe("User Auth", auth.UserAuth)
	_ = PDescribe("DP Auth", auth.DpAuth, Ordered)
	_ = PDescribe("Offline Auth", auth.OfflineAuth, Ordered)
	_ = PDescribe("Gateway", gateway.Gateway, Ordered)
	_ = PDescribe("Gateway - Cross-mesh", gateway.CrossMeshGatewayOnUniversal, Ordered)
	_ = PDescribe("HealthCheck panic threshold", healthcheck.HealthCheckPanicThreshold, Ordered)
	_ = PDescribe("HealthCheck", healthcheck.Policy)
	_ = PDescribe("MeshHealthCheck panic threshold", meshhealthcheck.MeshHealthCheckPanicThreshold, Ordered)
	_ = PDescribe("MeshHealthCheck", meshhealthcheck.MeshHealthCheck)
	_ = PDescribe("Service Probes", healthcheck.ServiceProbes, Ordered)
	_ = PDescribe("External Services", externalservices.Policy, Ordered)
	_ = PDescribe("External Services through Zone Egress", externalservices.ThroughZoneEgress, Ordered)
	_ = PDescribe("Inspect", inspect.Inspect, Ordered)
	_ = PDescribe("Applications Metrics", observability.ApplicationsMetrics, Ordered)
	_ = PDescribe("Tracing", observability.Tracing, Ordered)
	_ = PDescribe("MeshTrace", observability.PluginTest, Ordered)
	_ = PDescribe("Membership", membership.Membership, Ordered)
	_ = PDescribe("Traffic Logging", trafficlog.TCPLogging, Ordered)
	_ = PDescribe("MeshAccessLog", meshaccesslog.TestPlugin, Ordered)
	_ = PDescribe("Timeout", timeout.Policy, Ordered)
	_ = PDescribe("Retry", retry.Policy, Ordered)
	_ = PDescribe("MeshRetry", meshretry.HttpRetry, Ordered)
	_ = PDescribe("MeshRetry", meshretry.GrpcRetry, Ordered)
	_ = PDescribe("RateLimit", ratelimit.Policy, Ordered)
	_ = PDescribe("ProxyTemplate", proxytemplate.ProxyTemplate, Ordered)
	_ = PDescribe("MeshProxyPatch", meshproxypatch.MeshProxyPatch, Ordered)
	_ = PDescribe("Matching", matching.Matching, Ordered)
	_ = PDescribe("Mtls", mtls.Policy, Ordered)
	_ = PDescribe("Reachable Services", reachableservices.ReachableServices, Ordered)
	_ = PDescribe("Apis", api.Api, Ordered)
	_ = PDescribe("Traffic Permission", trafficpermission.TrafficPermissionUniversal, Ordered)
	_ = PDescribe("Traffic Route", trafficroute.TrafficRoute, Ordered)
	_ = PDescribe("Zone Egress", zoneegress.ExternalServices, Ordered)
	_ = PDescribe("Virtual Outbound", virtualoutbound.VirtualOutbound, Ordered)
	_ = PDescribe("Transparent Proxy", transparentproxy.TransparentProxy, Ordered)
	_ = FDescribe("Mesh Traffic Permission", meshtrafficpermission.MeshTrafficPermissionUniversal, Ordered)
	_ = PDescribe("GRPC", grpc.GRPC, Ordered)
	_ = PDescribe("MeshRateLimit", meshratelimit.Policy, Ordered)
	_ = PDescribe("MeshTimeout", timeout.PluginTest, Ordered)
	_ = PDescribe("Projected Service Account Token", projectedsatoken.ProjectedServiceAccountToken, Ordered)
	_ = PDescribe("Compatibility", compatibility.UniversalCompatibility, Label("arm-not-supported"), Ordered)
	_ = PDescribe("Resilience", resilience.ResilienceStandaloneUniversal, Ordered)
	_ = PDescribe("Leader Election", resilience.LeaderElectionPostgres, Ordered)
	_ = PDescribe("MeshFaultInjection", meshfaultinjection.Policy, Ordered)
)
