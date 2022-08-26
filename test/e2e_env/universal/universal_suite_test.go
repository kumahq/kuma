package auth_test

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e_env/universal/api"
	"github.com/kumahq/kuma/test/e2e_env/universal/auth"
	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	"github.com/kumahq/kuma/test/e2e_env/universal/externalservices"
	"github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	"github.com/kumahq/kuma/test/e2e_env/universal/healthcheck"
	"github.com/kumahq/kuma/test/e2e_env/universal/inspect"
	"github.com/kumahq/kuma/test/e2e_env/universal/matching"
	"github.com/kumahq/kuma/test/e2e_env/universal/membership"
	"github.com/kumahq/kuma/test/e2e_env/universal/mtls"
	"github.com/kumahq/kuma/test/e2e_env/universal/observability"
	"github.com/kumahq/kuma/test/e2e_env/universal/proxytemplate"
	"github.com/kumahq/kuma/test/e2e_env/universal/ratelimit"
	"github.com/kumahq/kuma/test/e2e_env/universal/reachableservices"
	"github.com/kumahq/kuma/test/e2e_env/universal/retry"
	"github.com/kumahq/kuma/test/e2e_env/universal/timeout"
	"github.com/kumahq/kuma/test/e2e_env/universal/trafficlog"
	"github.com/kumahq/kuma/test/e2e_env/universal/trafficpermission"
	"github.com/kumahq/kuma/test/e2e_env/universal/trafficroute"
	"github.com/kumahq/kuma/test/e2e_env/universal/transparentproxy"
	"github.com/kumahq/kuma/test/e2e_env/universal/virtualoutbound"
	"github.com/kumahq/kuma/test/e2e_env/universal/zoneegress"
	. "github.com/kumahq/kuma/test/framework"
)

func TestE2E(t *testing.T) {
	SetDefaultConsistentlyDuration(time.Second * 5)
	SetDefaultConsistentlyPollingInterval(time.Millisecond * 200)
	SetDefaultEventuallyPollingInterval(time.Millisecond * 500)
	SetDefaultEventuallyTimeout(time.Second * 10)
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
		Expect(env.Cluster.Install(EgressUniversal(func(zone string) (string, error) {
			return env.Cluster.GetKuma().GenerateZoneEgressToken("")
		}))).To(Succeed())
		state := UniversalNetworkingState{
			ZoneEgress: env.Cluster.GetZoneEgressNetworking(),
			KumaCp:     env.Cluster.GetKuma().(*UniversalControlPlane).Networking(),
		}
		bytes, err := json.Marshal(state)
		Expect(err).ToNot(HaveOccurred())
		return bytes
	},
	func(bytes []byte) {
		if env.Cluster != nil {
			return // cluster was already initiated with first function
		}
		state := UniversalNetworkingState{}
		Expect(json.Unmarshal(bytes, &state)).To(Succeed())
		env.Cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		E2EDeferCleanup(env.Cluster.DismissCluster) // clean up any containers if needed
		cp, err := NewUniversalControlPlane(
			env.Cluster.GetTesting(),
			core.Standalone,
			env.Cluster.Name(),
			env.Cluster.Verbose(),
			state.KumaCp,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(env.Cluster.AddNetworking(state.ZoneEgress, Config.ZoneEgressApp)).To(Succeed())
		env.Cluster.SetCp(cp)
	},
)

var _ = Describe("User Auth", auth.UserAuth)
var _ = Describe("DP Auth", auth.DpAuth, Ordered)
var _ = Describe("Gateway", gateway.Gateway, Ordered)
var _ = Describe("Gateway - Cross-mesh", gateway.CrossMeshGatewayOnUniversal, Ordered)
var _ = Describe("HealthCheck panic threshold", healthcheck.HealthCheckPanicThreshold, Ordered)
var _ = Describe("HealthCheck", healthcheck.Policy)
var _ = Describe("Service Probes", healthcheck.ServiceProbes, Ordered)
var _ = Describe("External Services", externalservices.Policy, Ordered)
var _ = Describe("Inspect", inspect.Inspect, Ordered)
var _ = Describe("Applications Metrics", observability.ApplicationsMetrics, Ordered)
var _ = Describe("Tracing", observability.Tracing, Ordered)
var _ = Describe("Membership", membership.Membership, Ordered)
var _ = Describe("Traffic Logging", trafficlog.TCPLogging, Ordered)
var _ = Describe("Timeout", timeout.Policy, Ordered)
var _ = Describe("Retry", retry.Policy, Ordered)
var _ = Describe("RateLimit", ratelimit.Policy, Ordered)
var _ = Describe("ProxyTemplate", proxytemplate.ProxyTemplate, Ordered)
var _ = Describe("Matching", matching.Matching, Ordered)
var _ = Describe("Mtls", mtls.Policy, Ordered)
var _ = Describe("Reachable Services", reachableservices.ReachableServices, Ordered)
var _ = Describe("Apis", api.Api, Ordered)
var _ = Describe("Traffic Permission", trafficpermission.TrafficPermissionUniversal, Ordered)
var _ = Describe("Traffic Route", trafficroute.TrafficRoute, Ordered)
var _ = Describe("Zone Egress", zoneegress.ExternalServices, Ordered)
var _ = Describe("Virtual Outbound", virtualoutbound.VirtualOutbound, Ordered)
var _ = Describe("Transparent Proxy", transparentproxy.TransparentProxy, Ordered)
