package auth_test

import (
	"encoding/json"
	"testing"
	"time"

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
	"github.com/kumahq/kuma/test/e2e_env/universal/matching"
	"github.com/kumahq/kuma/test/e2e_env/universal/membership"
	"github.com/kumahq/kuma/test/e2e_env/universal/observability"
	"github.com/kumahq/kuma/test/e2e_env/universal/proxytemplate"
	"github.com/kumahq/kuma/test/e2e_env/universal/ratelimit"
	"github.com/kumahq/kuma/test/e2e_env/universal/retry"
	"github.com/kumahq/kuma/test/e2e_env/universal/timeout"
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
		pf := env.Cluster.GetKuma().(*UniversalControlPlane).Networking()
		bytes, err := json.Marshal(pf)
		Expect(err).ToNot(HaveOccurred())
		return bytes
	},
	func(bytes []byte) {
		if env.Cluster != nil {
			return // cluster was already initiated with first function
		}
		networking := UniversalNetworking{}
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

var _ = Describe("User Auth", auth.UserAuth)
var _ = Describe("DP Auth", auth.DpAuth, Ordered)
var _ = Describe("Cross-mesh Gateway", gateway.CrossMeshGatewayOnUniversal, Ordered)
var _ = Describe("HealthCheck panic threshold", healthcheck.HealthCheckPanicThreshold, Ordered)
var _ = Describe("HealthCheck", healthcheck.Policy)
var _ = Describe("Service Probes", healthcheck.ServiceProbes, Ordered)
var _ = Describe("External Services", externalservices.Policy, Ordered)
var _ = Describe("Inspect", inspect.Inspect, Ordered)
var _ = Describe("Applications Metrics", observability.ApplicationsMetrics, Ordered)
var _ = Describe("Tracing", observability.Tracing, Ordered)
var _ = Describe("Membership", membership.Membership, Ordered)
var _ = Describe("Timeout", timeout.Policy, Ordered)
var _ = Describe("Retry", retry.Policy, Ordered)
var _ = Describe("RateLimit", ratelimit.Policy, Ordered)
var _ = Describe("ProxyTemplate", proxytemplate.ProxyTemplate, Ordered)
var _ = Describe("Matching", matching.Matching, Ordered)
