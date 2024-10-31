package externalservices

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

const nonDefaultMesh = "non-default"

func HybridUniversalGlobal() {
	meshMTLSOn := `
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
networking:
  outbound:
    passthrough: %s
routing:
  zoneEgress: %s
`

	externalService1 := `
type: ExternalService
mesh: %s
name: external-service-1
tags:
  kuma.io/service: external-service-1
  kuma.io/protocol: http
networking:
  address: es-test-server.default.svc.cluster.local:80
`

	externalService2 := `
type: ExternalService
mesh: %s
name: external-service-2
tags:
  kuma.io/service: external-service-2
  kuma.io/protocol: http
networking:
  address: "%s"
`

	// Override wait_for_warm_on_init to false because universal zone cannot resolve "es-test-server.default.svc.cluster.local:80"
	// The default (true) slows down ACK of all warming all the clusters delivered to universal client, even if only one cluster has problem.
	// This speeds up the test by at least 60s.
	ptWaitForWarmOnInit := `
type: ProxyTemplate
mesh: non-default
name: custom-template-1
selectors:
  - match:
      kuma.io/service: '*'
conf:
  imports:
    - default-proxy
  modifications:
    - cluster:
        operation: patch
        match:
          origin: outbound
        value: |
          wait_for_warm_on_init: false
`

	var global, zone1 Cluster
	var zone4 *UniversalCluster

	BeforeAll(func() {
		// Global
		global = NewUniversalCluster(NewTestingT(), Kuma5, Silent)

		Expect(NewClusterSetup().
			Install(Kuma(config_core.Global)).
			Install(YamlUniversal(fmt.Sprintf(meshMTLSOn, nonDefaultMesh, "true", "true"))).
			Install(MeshTrafficPermissionAllowAllUniversal(nonDefaultMesh)).
			Install(YamlUniversal(ptWaitForWarmOnInit)).
			Install(YamlUniversal(fmt.Sprintf(externalService1, nonDefaultMesh))).
			Setup(global)).To(Succeed())

		globalCP := global.GetKuma()

		group := errgroup.Group{}

		// K8s Cluster 1
		zone1 = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		NewClusterSetup().
			Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))). // do not deploy Egress
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(TestNamespace), democlient.WithMesh(nonDefaultMesh)),
				testserver.Install(
					testserver.WithName("es-test-server"),
					testserver.WithNamespace("default"),
					testserver.WithEchoArgs("echo", "--instance", "es-test-server"),
				),
			)).
			SetupInGroup(zone1, &group)

		// Universal Cluster 4
		zone4 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
		NewClusterSetup().
			Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))). // do not deploy Egress
			Install(IngressUniversal(global.GetKuma().GenerateZoneIngressToken)).
			Install(Parallel(
				DemoClientUniversal(
					"zone4-demo-client",
					nonDefaultMesh,
					WithTransparentProxy(true),
				),
				func(cluster Cluster) error {
					return cluster.DeployApp(
						WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", "es-test-server"}),
						WithName("es-test-server"),
						WithoutDataplane(),
						WithVerbose())
				},
				TestServerUniversal("test-server", nonDefaultMesh,
					WithArgs([]string{"echo", "--instance", "test-server"}),
					WithTransparentProxy(true),
				),
			)).
			SetupInGroup(zone4, &group)

		Expect(group.Wait()).To(Succeed())

		Expect(global.GetKumactlOptions().
			KumactlApplyFromString(
				fmt.Sprintf(externalService2, nonDefaultMesh, net.JoinHostPort(zone4.GetApp("es-test-server").GetIP(), "8080"))),
		).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
		Expect(zone4.DismissCluster()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("passthrough false with zoneegress false", func() {
		Expect(YamlUniversal(fmt.Sprintf(meshMTLSOn, nonDefaultMesh, "false", "false"))(global)).To(Succeed())

		By("reaching external service from k8s")
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				zone1, "demo-client", "external-service-1.mesh",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("es-test-server"))
		}, "30s", "1s").Should(Succeed())

		By("reaching external service from universal")
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				zone4, "zone4-demo-client", "external-service-2.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("es-test-server"))
		}, "30s", "1s").Should(Succeed())

		By("reaching internal service from k8s")
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				zone1, "demo-client", "test-server.mesh",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("test-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("passthrough false with zoneegress true", func() {
		Expect(YamlUniversal(fmt.Sprintf(meshMTLSOn, nonDefaultMesh, "false", "true"))(global)).To(Succeed())

		By("not reaching external service from k8s when zone egress is down")
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				zone1, "demo-client", "external-service-1.mesh",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(56), Equal(7), Equal(28)))
		}, "30s", "1s").Should(Succeed())

		By("not reaching external service from universal when zone egress is down")
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				zone4, "zone4-demo-client", "external-service-2.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(56), Equal(7), Equal(28)))
		}, "30s", "1s").Should(Succeed())

		By("not reaching internal service from k8s when zone egress is down")
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				zone1, "demo-client", "test-server.mesh",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(56), Equal(7), Equal(28)))
		}, "30s", "1s").Should(Succeed())
	})
}
