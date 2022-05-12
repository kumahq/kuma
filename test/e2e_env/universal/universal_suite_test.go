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
	"github.com/kumahq/kuma/test/e2e_env/universal/healthcheck"
	"github.com/kumahq/kuma/test/e2e_env/universal/inspect"
	"github.com/kumahq/kuma/test/e2e_env/universal/membership"
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

var _ = Describe("User Auth", auth.UserAuth)
var _ = Describe("DP Auth", auth.DpAuth, Ordered)
var _ = Describe("HealthCheck panic threshold", healthcheck.HealthCheckPanicThreshold, Ordered)
var _ = Describe("Service Probes", healthcheck.ServiceProbes, Ordered)
var _ = Describe("External Services", externalservices.ExternalServiceHostHeader, Ordered)
var _ = Describe("Inspect", inspect.Inspect, Ordered)
var _ = Describe("Membership", membership.Membership)
