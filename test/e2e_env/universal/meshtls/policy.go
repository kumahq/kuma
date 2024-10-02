package meshtls

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	var testServerContainerName string
	var testServer2ContainerName string
	meshName := "mesh-tls"
	testServerName := "mesh-tls-test-server"
	testServer2Name := "mesh-tls-test-server-2"

	BeforeAll(func() {
		testServerContainerName = fmt.Sprintf("%s_%s", universal.Cluster.Name(), testServerName)
		testServer2ContainerName = fmt.Sprintf("%s_%s", universal.Cluster.Name(), testServer2Name)
		Expect(NewClusterSetup().
			Install(ResourceUniversal(samples.MeshMTLSBuilder().WithName(meshName).Build())).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(TestServerUniversal(
				testServerName, meshName,
				WithArgs([]string{"echo", "--instance", "test-server"}),
				WithServiceName("mesh-tls-test-server"),
				WithDockerContainerName(testServerContainerName),
			)).
			Install(TestServerUniversal(
				testServer2Name, meshName,
				WithArgs([]string{"echo", "--instance", "test-server-2"}),
				WithServiceName("mesh-tls-test-server-2"),
				WithDockerContainerName(testServer2ContainerName),
			)).
			Install(DemoClientUniversal("mesh-tls-demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("mesh-tls-demo-client-no-mesh", "", WithoutDataplane())).
			Setup(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should change single dataplane to Permissive", func() {
		policy := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-policy
spec:
  targetRef:
    kind: MeshSubset
    tags:
      kuma.io/service: %s
  from:
    - targetRef:
        kind: Mesh
      default:
        mode: Permissive`, meshName, testServerName)
		// when
		// default strict mode on mesh
		Expect(universal.Cluster.Install(
			ResourceUniversal(samples.MeshMTLSBuilder().WithName(meshName).Build()),
		)).To(Succeed())

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server-2 from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server-2.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server-2"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// cannot access test-server from service outside of the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectFailure(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServerContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Exitcode).To(Equal(52))
		}, "30s", "500ms").Should(Succeed())

		// and
		// cannot access test-server-2 from service outside of the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectFailure(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServer2ContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Exitcode).To(Equal(52))
		}, "30s", "500ms").Should(Succeed())

		// when
		// applied MeshTLS policy to set Permissive mode on test-server
		Expect(universal.Cluster.Install(YamlUniversal(policy))).To(Succeed())

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server-2 from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server-2.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server-2"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server from service outside of the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServerContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// cannot access test-server-2 from service outside of the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectFailure(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServer2ContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Exitcode).To(Equal(52))
		}, "30s", "500ms").Should(Succeed())
	})

	It("should change single dataplane to Strict", func() {
		policy := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-policy
spec:
  targetRef:
    kind: MeshSubset
    tags:
      kuma.io/service: %s
  from:
    - targetRef:
        kind: Mesh
      default:
        mode: Strict`, meshName, testServerName)
		// when
		// default strict mode on mesh
		Expect(universal.Cluster.Install(
			ResourceUniversal(samples.MeshMTLSBuilder().WithPermissiveMTLSBackends().WithName(meshName).Build()),
		)).To(Succeed())

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server-2 from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server-2.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server-2"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server from service outside of the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServerContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server-2 from service outside of the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServer2ContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server-2"))
		}, "30s", "500ms").Should(Succeed())

		// when
		// applied MeshTLS policy to set Strict mode on test-server
		Expect(universal.Cluster.Install(YamlUniversal(policy))).To(Succeed())

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server-2 from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server-2.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server-2"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server from service outside of the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectFailure(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServerContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Exitcode).To(Equal(52))
		}, "30s", "500ms").Should(Succeed())

		// and
		// cannot access test-server-2 from service outside of the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServer2ContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server-2"))
		}, "30s", "500ms").Should(Succeed())
	})

	It("should tls version for 1.3", func() {
		// given
		admin, err := universal.Cluster.GetApp(testServerName).GetEnvoyAdminTunnel()
		Expect(err).ToNot(HaveOccurred())

		policy := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-policy
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        tlsVersion:
          min: TLS13
          max: TLS13`, meshName)
		// when
		// default strict mode on mesh
		Expect(universal.Cluster.Install(
			ResourceUniversal(samples.MeshMTLSBuilder().WithName(meshName).Build()),
		)).To(Succeed())

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		// and
		// uses tls version 1.2
		Eventually(func(g Gomega) {
			s, err := admin.GetStats("listener.(.*)_80.ssl.versions.TLSv1.2")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// when
		// applied MeshTLS policy to set 1.3 version on test-server
		Expect(universal.Cluster.Install(YamlUniversal(policy))).To(Succeed())

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// and
		// uses tls version 1.3
		Eventually(func(g Gomega) {
			s, err := admin.GetStats("listener.(.*)_80.ssl.versions.TLSv1.3")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())
	})

	It("should set cypher and version", func() {
		// given
		admin, err := universal.Cluster.GetApp(testServerName).GetEnvoyAdminTunnel()
		Expect(err).ToNot(HaveOccurred())

		policy := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-policy
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        tlsVersion:
          min: TLS12
          max: TLS12
        tlsCiphers:
        - "ECDHE-RSA-AES256-GCM-SHA384"`, meshName)
		// when
		// default strict mode on mesh
		Expect(universal.Cluster.Install(
			ResourceUniversal(samples.MeshMTLSBuilder().WithName(meshName).Build()),
		)).To(Succeed())

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())

		// and
		// uses tls version 1.2
		Eventually(func(g Gomega) {
			s, err := admin.GetStats("listener.(.*)_80.ssl.versions.TLSv1.2")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// and
		// doesn't use specified cypher
		Eventually(func(g Gomega) {
			s, err := admin.GetStats("listener.(.*)_80.ssl.ciphers.ECDHE-RSA-AES256-GCM-SHA384")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "30s", "1s").Should(Succeed())

		// when
		// applied MeshTLS policy to set cypher on test-server
		Expect(universal.Cluster.Install(YamlUniversal(policy))).To(Succeed())

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// and
		// uses tls version 1.2
		Eventually(func(g Gomega) {
			s, err := admin.GetStats("listener.(.*)_80.ssl.versions.TLSv1.2")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// and
		// doesn't uses specific cypher
		Eventually(func(g Gomega) {
			s, err := admin.GetStats("listener.(.*)_80.ssl.ciphers.ECDHE-RSA-AES256-GCM-SHA384")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())
	})
}
