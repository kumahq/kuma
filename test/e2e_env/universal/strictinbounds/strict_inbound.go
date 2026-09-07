package strictinbounds

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
)

func StrictInboundPorts() {
	const meshName = "strict-inbound-ports"
	const identityName = "strict-inbound-ports-identity"

	// enableMTLS gives the proxies a workload identity, which is what turns
	// their inbounds into mTLS listeners.
	enableMTLS := func() InstallFunc {
		return Combine(
			MeshIdentityBundled(meshName, identityName),
			MeshTrafficPermissionAllowAllUniversalWorkloadIdentity(
				meshName,
				MeshIdentityTrustDomain(meshName, universal.Cluster),
			),
		)
	}

	permissiveMeshTLS := fmt.Sprintf(`
type: MeshTLS
name: permissive
mesh: %s
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        mode: Permissive
`, meshName)

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

	BeforeEach(func() {
		err := NewClusterSetup().
			Install(Yaml(samples.MeshDefaultBuilder().
				WithName(meshName),
			)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterEach(func() {
		// drop the identity again, so the next case starts without mTLS
		for _, policy := range []struct{ plural, singular string }{
			{"meshtlses", "meshtls"},
			{"meshtrafficpermissions", "meshtrafficpermission"},
			{"meshidentities", "meshidentity"},
		} {
			items, err := universal.Cluster.GetKumactlOptions().KumactlList(policy.plural, meshName)
			Expect(err).ToNot(HaveOccurred())
			for _, item := range items {
				Expect(universal.Cluster.GetKumactlOptions().KumactlDelete(policy.singular, item, meshName)).To(Succeed())
			}
		}
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
				universal.Cluster, "demo-client", "test-server.svc.mesh.local",
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

		// and
		// not secured DPP can be accessed
		notSecuredDPPInboundAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-not-secure").GetIP(), "80")
		notSecuredServiceAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-not-secure").GetIP(), "8080")
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server-not-secure.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server-not-secure"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", notSecuredDPPInboundAddress,
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
	})

	It("should allow all traffic when permissive mode", func() {
		err := NewClusterSetup().
			Install(YamlUniversal(permissiveMeshTLS)).
			Install(enableMTLS()).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		dppInboundAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		serviceAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "8080")

		// then communication should works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local",
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

		// and
		// not secured DPP can be accessed
		notSecuredDPPInboundAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-not-secure").GetIP(), "80")
		notSecuredServiceAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-not-secure").GetIP(), "8080")
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server-not-secure.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server-not-secure"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client-not-in-mesh", notSecuredDPPInboundAddress,
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
	})

	It("should allow only traffic to specific ports when strict mode", func() {
		err := NewClusterSetup().
			Install(enableMTLS()).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		dppInboundAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		serviceAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "8080")

		// then
		// communication should works only to DPP port
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		// and
		// the service port cannot be accessed
		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				universal.Cluster, "demo-client-not-in-mesh", dppInboundAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Exitcode).To(Or(Equal(52)))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				universal.Cluster, "demo-client", serviceAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Exitcode).To(Or(Equal(52)))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				universal.Cluster, "demo-client-not-in-mesh", serviceAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			// 52 = CURLE_GOT_NOTHING (TLS alert, clean close)
			// 56 = CURLE_RECV_ERROR (TCP RST when no filter chain matches)
			g.Expect(resp.Exitcode).To(Or(Equal(52), Equal(56)))
		}, "30s", "1s").Should(Succeed())

		// and
		// not secured DPP can be accessed
		notSecuredDPPInboundAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-not-secure").GetIP(), "80")
		notSecuredServiceAddress := net.JoinHostPort(universal.Cluster.GetApp("test-server-not-secure").GetIP(), "8080")

		// then
		// communication should works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server-not-secure.svc.mesh.local",
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
			resp, err := client.CollectFailure(
				universal.Cluster, "demo-client-not-in-mesh", notSecuredDPPInboundAddress,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Exitcode).To(Or(Equal(52)))
		}, "30s", "1s").Should(Succeed())
	})
}
