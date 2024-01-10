package healthcheck

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func ServiceProbes() {
	const meshName = "service-probes"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TestServerUniversal("test-server", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"}),
				ProxyOnly(),
				ServiceProbe()),
			).
			Install(TrafficRouteUniversal(meshName)).
			Install(TrafficPermissionUniversal(meshName)).
			Install(DemoClientUniversal("demo-client", meshName, ServiceProbe())).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	ProxyHealthy := func(name string) {
		Eventually(func() (string, error) {
			return universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplane", name, "-oyaml", "--mesh", meshName)
		}, "30s", "1s").Should(ContainSubstring("ready: true"))
	}

	ProxyUnhealthy := func(name string) {
		Eventually(func() (string, error) {
			return universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplane", name, "-oyaml", "--mesh", meshName)
		}, "30s", "1s").Should(ContainSubstring("health: {}"))
	}

	It("should update dataplane.inbound.health of unhealthy test-server", func() {
		ProxyUnhealthy("test-server")
	})

	It("should mark DP as unhealthy when listeners are draining", func() {
		// given
		ProxyHealthy("demo-client")

		// when kuma-dp is draining state (not terminated) after receiving with SIGTERM
		_, _, err := universal.Cluster.Exec("", "", "demo-client", "pkill", "-15", "kuma-dp")
		Expect(err).ToNot(HaveOccurred())

		// then proxy is offline rather than missing
		ProxyUnhealthy("demo-client")
	})
}
