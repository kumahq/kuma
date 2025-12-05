package strictinbounds

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func StrictInboundPorts() {
	const meshName = "strict-inbound-ports"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-not-in-mesh", "", WithoutDataplane())).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "test-server"}))).
			Install(TestServerUniversal("test-server-not-secure", meshName,
				WithServiceName("test-server-not-secure"),
				WithDpEnvs(map[string]string{
					"KUMA_DATAPLANE_RUNTIME_STRICT_INBOUND_PORTS_ENABLED": "false",
				}),
				WithArgs([]string{"echo", "--instance", "test-server-not-secure"}))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should allow all traffic when there is no tls", func() {
		dppInboundAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		serviceAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "8080")

		// then communication should works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", dppInboundAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", serviceAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", serviceAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should allow all traffic when permissive mode", func() {
		err := NewClusterSetup().
			Install(Yaml(samples.MeshDefaultBuilder().
				WithName(meshName).
				WithBuiltinMTLSBackend("backend").
				WithEnabledMTLSBackend("backend").
				WithPermissiveMTLSBackends(),
			)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		dppInboundAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		serviceAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "8080")

		// then communication should works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", dppInboundAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", serviceAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", serviceAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should allow only traffic to specific ports when strict mode", func() {
		err := NewClusterSetup().
			Install(Yaml(samples.MeshDefaultBuilder().
				WithName(meshName).
				WithBuiltinMTLSBackend("backend").
				WithEnabledMTLSBackend("backend"),
			)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		dppInboundAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		serviceAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "8080")

		// then
		// communication should works only to DPP port
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		// and
		// the service port cannot be accessed
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", dppInboundAddress,
			)
			g.Expect(err).To(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", serviceAddress,
			)
			g.Expect(err).To(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", serviceAddress,
			)
			g.Expect(err).To(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		// and
		// not secured DPP can be accessed
		notSecuredDPPInboundAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-not-secure").GetIP(), "80")
		notSecuredServiceAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-not-secure").GetIP(), "8080")

		// then
		// communication should works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server-not-secure.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server-not-secure"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", notSecuredServiceAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server-not-secure"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", notSecuredServiceAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server-not-secure"))
		}, "30s", "1s").Should(Succeed())

		// and
		// the dpp port cannot be accessed outside of the mesh
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", notSecuredDPPInboundAddress,
			)
			g.Expect(err).To(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
