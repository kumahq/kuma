package auth_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e_env/universal/auth"
	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	"github.com/kumahq/kuma/test/e2e_env/universal/externalservices"
	"github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	"github.com/kumahq/kuma/test/e2e_env/universal/healthcheck"
	"github.com/kumahq/kuma/test/e2e_env/universal/inspect"
<<<<<<< HEAD
=======
	"github.com/kumahq/kuma/test/e2e_env/universal/intercp"
	"github.com/kumahq/kuma/test/e2e_env/universal/matching"
>>>>>>> fdc88c788 (fix(kuma-cp): add components in runtime (#6350))
	"github.com/kumahq/kuma/test/e2e_env/universal/membership"
	"github.com/kumahq/kuma/test/e2e_env/universal/metrics"
	. "github.com/kumahq/kuma/test/framework"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Universal Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		env.Cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		E2EDeferCleanup(env.Cluster.DismissCluster)
		Expect(env.Cluster.Install(Kuma(core.Standalone,
			WithEnv("KUMA_STORE_UNSAFE_DELETE", "true"),
			WithEnv("KUMA_XDS_SERVER_DATAPLANE_STATUS_FLUSH_INTERVAL", "1s"), // speed up some tests by flushing stats quicker than default 10s
		))).To(Succeed())
		pf := env.Cluster.GetKuma().(*UniversalControlPlane).Networking()
		bytes, err := json.Marshal(pf)
		Expect(err).ToNot(HaveOccurred())
		return bytes
	},
	func(bytes []byte) {
		if env.Cluster != nil {
			return // cluster was already initiated with first function
		}
		networking := UniversalCPNetworking{}
		Expect(json.Unmarshal(bytes, &networking)).To(Succeed())
		env.Cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		E2EDeferCleanup(env.Cluster.DismissCluster) // clean up any containers if needed
		cp, err := NewUniversalControlPlane(
			env.Cluster.GetTesting(),
			core.Standalone,
			env.Cluster.Name(),
			env.Cluster.Verbose(),
			networking,
		)
		Expect(err).ToNot(HaveOccurred())
		env.Cluster.SetCp(cp)
	},
)

<<<<<<< HEAD
var _ = Describe("User Auth", auth.UserAuth)
var _ = Describe("DP Auth", auth.DpAuth, Ordered)
var _ = Describe("Cross-mesh Gateway", gateway.CrossMeshGatewayOnUniversal, Ordered)
var _ = Describe("HealthCheck panic threshold", healthcheck.HealthCheckPanicThreshold, Ordered)
var _ = Describe("HealthCheck", healthcheck.Policy)
var _ = Describe("Service Probes", healthcheck.ServiceProbes, Ordered)
var _ = Describe("External Services", externalservices.ExternalServiceHostHeader, Ordered)
var _ = Describe("Inspect", inspect.Inspect, Ordered)
var _ = Describe("Applications Metrics", metrics.ApplicationsMetrics, Ordered)
var _ = Describe("Membership", membership.Membership, Ordered)
=======
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
	_ = Describe("Traffic Permission", trafficpermission.TrafficPermissionUniversal, Ordered)
	_ = Describe("Traffic Route", trafficroute.TrafficRoute, Ordered)
	_ = Describe("Zone Egress", zoneegress.ExternalServices, Ordered)
	_ = Describe("Virtual Outbound", virtualoutbound.VirtualOutbound, Ordered)
	_ = Describe("Transparent Proxy", transparentproxy.TransparentProxy, Ordered)
	_ = Describe("Mesh Traffic Permission", meshtrafficpermission.MeshTrafficPermissionUniversal, Ordered)
	_ = Describe("GRPC", grpc.GRPC, Ordered)
	_ = Describe("MeshRateLimit", meshratelimit.Policy, Ordered)
	_ = Describe("MeshTimeout", timeout.PluginTest, Ordered)
	_ = Describe("Projected Service Account Token", projectedsatoken.ProjectedServiceAccountToken, Ordered)
	_ = Describe("Compatibility", compatibility.UniversalCompatibility, Label("arm-not-supported"), Ordered)
	_ = Describe("Resilience", resilience.ResilienceStandaloneUniversal, Ordered)
	_ = Describe("Leader Election", resilience.LeaderElectionPostgres, Ordered)
	_ = Describe("MeshFaultInjection", meshfaultinjection.Policy, Ordered)
	_ = Describe("MeshLoadBalancingStrategy", meshloadbalancingstrategy.Policy, Ordered)
	_ = Describe("InterCP Server", intercp.InterCP, Ordered)
)
>>>>>>> fdc88c788 (fix(kuma-cp): add components in runtime (#6350))
