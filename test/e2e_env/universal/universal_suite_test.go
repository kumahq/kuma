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

var _ = SynchronizedBeforeSuite(func() []byte { return universal.SetupAndGetState() }, universal.RestoreState)

var _ = Describe("User Auth", auth.UserAuth)
var _ = Describe("DP Auth", auth.DpAuth, Ordered)
var _ = Describe("Gateway", gateway.Gateway, Ordered)
var _ = Describe("Gateway - Cross-mesh", gateway.CrossMeshGatewayOnUniversal, Ordered)
var _ = Describe("HealthCheck panic threshold", healthcheck.HealthCheckPanicThreshold, Ordered)
var _ = Describe("HealthCheck", healthcheck.Policy)
var _ = Describe("MeshHealthCheck panic threshold", meshhealthcheck.MeshHealthCheckPanicThreshold, Ordered)
var _ = Describe("MeshHealthCheck", meshhealthcheck.MeshHealthCheck)
var _ = Describe("Service Probes", healthcheck.ServiceProbes, Ordered)
var _ = Describe("External Services", externalservices.Policy, Ordered)
var _ = Describe("External Services through Zone Egress", externalservices.ThroughZoneEgress, Ordered)
var _ = Describe("Inspect", inspect.Inspect, Ordered)
var _ = Describe("Applications Metrics", observability.ApplicationsMetrics, Ordered)
var _ = Describe("Tracing", observability.Tracing, Ordered)
var _ = Describe("MeshTrace", observability.PluginTest, Ordered)
var _ = Describe("Membership", membership.Membership, Ordered)
var _ = Describe("Traffic Logging", trafficlog.TCPLogging, Ordered)
var _ = Describe("MeshAccessLog", meshaccesslog.TestPlugin, Ordered)
var _ = Describe("Timeout", timeout.Policy, Ordered)
var _ = Describe("Retry", retry.Policy, Ordered)
var _ = Describe("MeshRetry", meshretry.HttpRetry, Ordered)
var _ = Describe("MeshRetry", meshretry.GrpcRetry, Ordered)
var _ = Describe("RateLimit", ratelimit.Policy, Ordered)
var _ = Describe("ProxyTemplate", proxytemplate.ProxyTemplate, Ordered)
var _ = Describe("MeshProxyPatch", meshproxypatch.MeshProxyPatch, Ordered)
var _ = Describe("Matching", matching.Matching, Ordered)
var _ = Describe("Mtls", mtls.Policy, Ordered)
var _ = Describe("Reachable Services", reachableservices.ReachableServices, Ordered)
var _ = Describe("Apis", api.Api, Ordered)
var _ = Describe("Traffic Permission", trafficpermission.TrafficPermissionUniversal, Ordered)
var _ = Describe("Traffic Route", trafficroute.TrafficRoute, Ordered)
var _ = Describe("Zone Egress", zoneegress.ExternalServices, Ordered)
var _ = Describe("Virtual Outbound", virtualoutbound.VirtualOutbound, Ordered)
var _ = Describe("Transparent Proxy", transparentproxy.TransparentProxy, Ordered)
var _ = Describe("Mesh Traffic Permission", meshtrafficpermission.MeshTrafficPermissionUniversal, Ordered)
var _ = Describe("GRPC", grpc.GRPC, Ordered)
var _ = Describe("MeshRateLimit", meshratelimit.Policy, Ordered)
var _ = Describe("MeshTimeout", timeout.PluginTest, Ordered)
var _ = Describe("Projected Service Account Token", projectedsatoken.ProjectedServiceAccountToken, Ordered)
var _ = Describe("Compatibility", compatibility.UniversalCompatibility, Label("arm-not-supported"), Ordered)
var _ = Describe("Resilience", resilience.ResilienceStandaloneUniversal, Ordered)
var _ = Describe("Leader Election", resilience.LeaderElectionPostgres, Ordered)
var _ = Describe("MeshFaultInjection", meshfaultinjection.Policy, Ordered)
