package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/api"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/auth"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/bindoutbounds"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/compatibility"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/envoyconfig"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/grpc"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/healthcheck"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/inspect"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/intercp"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshaccesslog"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshexternalservice"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshfaultinjection"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshhealthcheck"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshidentity"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshloadbalancingstrategy"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshproxypatch"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshratelimit"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshretry"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshservice"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshservicelabelpropagation"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshtls"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/meshtrafficpermission"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/mtls"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/observability"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/projectedsatoken"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/reachableservices"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/resilience"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/strictinbounds"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/timeout"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/transparentproxy"
	"github.com/kumahq/kuma/v3/test/e2e_env/universal/workload"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Universal Suite")
}

var (
	_ = E2ESynchronizedBeforeSuite(universal.SetupAndGetState, universal.RestoreState)
	_ = SynchronizedAfterSuite(func() {}, universal.SynchronizedAfterSuite)
	_ = ReportAfterSuite("universal after suite", universal.AfterSuite)
	// Opt-in (KUMA3_PREFLIGHT_BIN + KUMA3_PREFLIGHT_DIR): snapshot the shared CP after
	// each spec so kuma3-preflight can classify which tests use Kuma-3.0-removed
	// features. No-op otherwise. Snapshot filenames carry the parallel-process index,
	// so concurrent processes sharing the output dir do not collide.
	_ = AfterEach(func() {
		CapturePreflightCluster(CurrentSpecReport().FullText(), universal.Cluster)
	})
)

var (
	_ = Describe("User Auth", Label("job-3"), auth.UserAuth)
	_ = Describe("DP Auth", Label("job-3"), auth.DpAuth, Ordered)
	_ = Describe("Offline Auth", Label("job-3"), auth.OfflineAuth, Ordered)
	_ = Describe("MeshHealthCheck panic threshold", Label("job-1"), meshhealthcheck.MeshHealthCheckPanicThreshold, Ordered)
	_ = Describe("MeshHealthCheck", Label("job-1"), meshhealthcheck.MeshHealthCheck)
	_ = Describe("Workload", Label("job-3"), workload.Workload, Ordered)
	_ = Describe("Service Probes", Label("job-3"), healthcheck.ServiceProbes, Ordered)
	_ = Describe("Inspect", Label("job-3"), inspect.Inspect, Ordered)
	_ = Describe("Mesh External Services", Label("job-1"), meshexternalservice.MeshExternalService, Ordered)
	_ = Describe("MeshService", Label("job-3"), meshservice.MeshService, Ordered)
	_ = Describe("MeshService Label Propagation", Label("job-3"), meshservicelabelpropagation.LabelPropagation, Ordered)
	_ = Describe("MeshTrace", Label("job-3"), observability.PluginTest, Ordered)
	_ = Describe("MeshAccessLog", Label("job-1"), meshaccesslog.TestPlugin, Ordered)
	_ = Describe("MeshAccessLog - matches", Label("job-1"), meshaccesslog.Matches, Ordered)
	_ = Describe("MeshRetry", Label("job-2"), meshretry.HttpRetry, Ordered)
	_ = Describe("MeshRetry", Label("job-2"), meshretry.GrpcRetry, Ordered)
	_ = Describe("MeshProxyPatch", Label("job-3"), meshproxypatch.MeshProxyPatch, Ordered)
	_ = Describe("MeshProxyPatch on Zone Proxy", Label("job-3"), meshproxypatch.ZoneProxy, Ordered)
	_ = Describe("Mtls", Label("job-1"), mtls.Policy, Ordered)
	_ = Describe("Reachable Services", Label("job-3"), reachableservices.ReachableServices, Ordered)
	_ = Describe("Apis", Label("job-3"), api.Api, Ordered)
	_ = Describe("Transparent Proxy", Label("job-3"), transparentproxy.TransparentProxy, Ordered)
	_ = Describe("Mesh Traffic Permission", Label("job-2"), meshtrafficpermission.MeshTrafficPermissionUniversal, Ordered)
	_ = Describe("GRPC", Label("job-3"), grpc.GRPC, Ordered)
	_ = Describe("MeshRateLimit", Label("job-2"), meshratelimit.Policy, Ordered)
	_ = Describe("MeshRateLimit on Zone Proxy", Label("job-2"), meshratelimit.ZoneProxy, Ordered)
	_ = Describe("MeshTimeout", Label("job-3"), timeout.PluginTest, Ordered)
	_ = Describe("Projected Service Account Token", Label("job-3"), projectedsatoken.ProjectedServiceAccountToken, Ordered)
	_ = Describe("Compatibility", Label("job-2"), compatibility.UniversalCompatibility, Ordered)
	_ = Describe("Resilience", Label("job-2"), resilience.ResilienceUniversal, Ordered)
	_ = Describe("Leader Election", Label("job-2"), resilience.LeaderElectionPostgres, Ordered)
	_ = Describe("MeshFaultInjection", Label("job-2"), meshfaultinjection.Policy, Ordered)
	_ = Describe("MeshFaultInjection on Zone Proxy", Label("job-2"), meshfaultinjection.ZoneProxy, Ordered)
	_ = Describe("MeshLoadBalancingStrategy", Label("job-3"), meshloadbalancingstrategy.Policy, Ordered)
	_ = Describe("InterCP Server", Label("job-3"), intercp.InterCP, Ordered)
	_ = Describe("MeshTLS", Label("job-3"), meshtls.Policy, Ordered)
	_ = Describe("Envoy Config – Sidecars", Label("job-0"), Label("golden-files-e2e"), envoyconfig.Sidecars, Ordered)
	_ = Describe("Envoy Config – Zone Proxies", Label("job-0"), Label("golden-files-e2e"), envoyconfig.ZoneProxies, Ordered)
	_ = Describe("Bind Outbounds", Label("job-3"), Label("ipv6-not-supported"), bindoutbounds.BindToLoopbackAddresses, Ordered)
	_ = Describe("MeshIdentity - Spire", Label("job-2"), meshidentity.Spire, Ordered)
	_ = Describe("MeshIdentity - Rotate CA", Label("job-2"), meshidentity.Rotate, Ordered)
	_ = Describe("Strict Inbound Ports", Label("job-3"), strictinbounds.StrictInboundPorts, Ordered)
)
