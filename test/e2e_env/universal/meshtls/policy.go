package meshtls

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
)

func Policy() {
	var testServerContainerName string
	var testServer2ContainerName string
	meshName := "mesh-tls"
	identityName := "mesh-tls-identity"
	testServerName := "mesh-tls-test-server"
	testServer2Name := "mesh-tls-test-server-2"

	// Mesh-wide permissive mode, the baseline the Strict case starts from.
	meshPermissive := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-mesh-default
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        mode: Permissive`, meshName)

	BeforeAll(func() {
		testServerContainerName = fmt.Sprintf("%s_%s", universal.Cluster.Name(), testServerName)
		testServer2ContainerName = fmt.Sprintf("%s_%s", universal.Cluster.Name(), testServer2Name)
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(MeshIdentityBundled(meshName, identityName)).
			Install(MeshTrafficPermissionAllowAllUniversalWorkloadIdentity(
				meshName,
				MeshIdentityTrustDomain(meshName, universal.Cluster),
			)).
			Install(TestServerUniversal(
				testServerName, meshName,
				WithArgs([]string{"echo", "--instance", "test-server"}),
				WithServiceName("mesh-tls-test-server"),
				WithDockerContainerName(testServerContainerName),
				WithLabels(map[string]string{"kuma.io/display-name": "mesh-tls-test-server"}),
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

	E2EAfterEach(func() {
		// Every case starts from the default strict mode. Each one names its
		// MeshTLS differently on purpose: the mesh hash the control plane uses
		// to decide whether to regenerate a proxy's config covers a resource's
		// name and version but not its spec, and the universal store hands a
		// re-created resource version 1 again. Reusing one name across cases
		// therefore lets a later spec hash the same as an earlier one, and the
		// proxy keeps serving the earlier config.
		items, err := universal.Cluster.GetKumactlOptions().KumactlList("meshtlses", meshName)
		Expect(err).ToNot(HaveOccurred())
		for _, item := range items {
			Expect(universal.Cluster.GetKumactlOptions().KumactlDelete("meshtls", item, meshName)).To(Succeed())
		}
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should change single dataplane to Permissive", func() {
		policy := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-dpp-permissive
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/display-name: %s
  rules:
    - default:
        mode: Permissive`, meshName, testServerName)
		// given the mesh in its default strict mode

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server-2 from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server-2.svc.mesh.local",
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
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))

			// and
			// can access test-server-2 from service in the mesh
			responses, err = client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server-2.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server-2"))

			// and
			// can access test-server from service outside of the mesh
			responses, err = client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServerContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))

			// and
			// cannot access test-server-2 from service outside of the mesh
			failureResponses, err := client.CollectFailure(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServer2ContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(failureResponses.Exitcode).To(Equal(52))
		}, "30s", "500ms").Should(Succeed())
	})

	It("should change single dataplane to Strict", func() {
		policy := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-dpp-strict
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/display-name: %s
  rules:
    - default:
        mode: Strict`, meshName, testServerName)
		// when
		// permissive mode on the whole mesh
		Expect(universal.Cluster.Install(YamlUniversal(meshPermissive))).To(Succeed())

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))
		}, "30s", "500ms").Should(Succeed())

		// and
		// can access test-server-2 from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server-2.svc.mesh.local",
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
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))

			// and
			// can access test-server-2 from service in the mesh
			responses, err = client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server-2.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server-2"))

			// and
			// can access test-server from service outside of the mesh
			failureResponses, err := client.CollectFailure(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServerContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(failureResponses.Exitcode).To(Equal(52))

			// and
			// cannot access test-server-2 from service outside of the mesh
			responses, err = client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client-no-mesh", testServer2ContainerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server-2"))
		}, "30s", "500ms").Should(Succeed())
	})

	It("should tls version for 1.3", func() {
		// given
		admin := universal.Cluster.GetApp(testServerName).GetEnvoyAdminTunnel()

		policy := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-version-13
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        tlsVersion:
          min: TLS13
          max: TLS13`, meshName)
		// given the mesh in its default strict mode

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.svc.mesh.local",
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
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))

			// and
			// uses tls version 1.3
			s, err := admin.GetStats("listener.(.*)_80.ssl.versions.TLSv1.3")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s).To(stats.BeGreaterThanZero())
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())
	})

	It("should set cypher and version", func() {
		// given
		admin := universal.Cluster.GetApp(testServerName).GetEnvoyAdminTunnel()

		policy := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-version-12-ciphers
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        tlsVersion:
          min: TLS12
          max: TLS12
        tlsCiphers:
        - "ECDHE-RSA-AES256-GCM-SHA384"`, meshName)
		// given the mesh in its default strict mode

		// then
		// can access test-server from service in the mesh
		Eventually(func(g Gomega) {
			responses, err := client.CollectEchoResponse(
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.svc.mesh.local",
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
				universal.Cluster, "mesh-tls-demo-client", "mesh-tls-test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses.Instance).To(Equal("test-server"))

			// and
			// uses tls version 1.2
			s, err := admin.GetStats("listener.(.*)_80.ssl.versions.TLSv1.2")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s).To(stats.BeGreaterThanZero())

			// and
			// uses specific cypher
			s, err = admin.GetStats("listener.(.*)_80.ssl.ciphers.ECDHE-RSA-AES256-GCM-SHA384")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s).To(stats.BeGreaterThanZero())
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())
	})
}
