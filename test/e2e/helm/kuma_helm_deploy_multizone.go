package helm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
)

// meshZoneIngressApp is the Deployment, Service and MeshZoneAddress the chart
// names after the mesh it deploys the zone ingress for:
// '<chart name>-<mesh>-ingress'.
const meshZoneIngressApp = "kuma-default-ingress"

// ZoneWithHelmChartAndUniversalGlobal verifies the Helm chart's multizone Zone
// install path against a Universal Global, the only supported Global topology
// now that Global-on-Kubernetes has been removed (see #17270, #17272).
func ZoneWithHelmChartAndUniversalGlobal() {
	var globalCluster, zoneCluster Cluster

	BeforeAll(func() {
		globalCluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)
		zoneCluster = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueID()),
		)

		err := NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		global := globalCluster.GetKuma()
		Expect(global).ToNot(BeNil())

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithGlobalAddress(global.GetKDSServerAddress()),
				WithHelmOpt("meshes[0].name", "default"),
				WithHelmOpt("meshes[0].ingress.enabled", "true"),
				// The chart defaults the ingress Service to LoadBalancer, which
				// never gets an address on the k3d clusters the E2E tests run on.
				WithHelmOpt("meshes[0].ingress.service.type", "NodePort"),
			)).
			Install(WaitNumPods(Config.KumaNamespace, 1, meshZoneIngressApp)).
			Install(WaitPodsAvailable(Config.KumaNamespace, meshZoneIngressApp)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(TestNamespace), democlient.WithMesh("default")),
				testserver.Install(),
			)).
			Setup(zoneCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(globalCluster, "default")
		DebugKube(zoneCluster, "default", TestNamespace)
	})

	E2EAfterAll(func() {
		ControlPlaneAssertions(globalCluster)
		ControlPlaneAssertions(zoneCluster)
		Expect(zoneCluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zoneCluster.DeleteKuma()).To(Succeed())
		Expect(globalCluster.DeleteKuma()).To(Succeed())
		Expect(globalCluster.DismissCluster()).To(Succeed())
		Expect(zoneCluster.DismissCluster()).To(Succeed())
	})

	It("should deploy Zone via Helm chart against a Universal Global", func() {
		// mesh is synced to zone
		Eventually(func(g Gomega) {
			output, err := zoneCluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).To(ContainSubstring("default"))
		}, "5s", "500ms").Should(Succeed())

		// and dataplanes are synced to global
		Eventually(func(g Gomega) {
			out, _, err := globalCluster.GetKuma().Exec("curl", "--fail", "--show-error", "http://localhost:5681/dataplanes")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).Should(ContainSubstring("demo-client"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should register the chart's mesh-scoped zone ingress", func() {
		// The zone ingress is a regular Dataplane now, so it shows up among the
		// zone's dataplanes rather than as a ZoneIngress resource.
		Eventually(func(g Gomega) {
			dataplanes, err := zoneCluster.GetKumactlOptions().KumactlList("dataplanes", "default")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dataplanes).To(ContainElement(ContainSubstring(meshZoneIngressApp)))
		}, "60s", "1s").Should(Succeed())

		// MeshZoneAddress is what the other zones route to, and the controller
		// only publishes it once the ingress Service has a ready endpoint.
		Eventually(func(g Gomega) {
			_, err := k8s.RunKubectlAndGetOutputContextE(
				zoneCluster.GetTesting(), context.Background(),
				zoneCluster.GetKubectlOptions(Config.KumaNamespace),
				"get", "meshzoneaddress", meshZoneIngressApp,
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "60s", "1s").Should(Succeed())
	})

	It("communication in between apps in zone works", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(zoneCluster, "demo-client", "http://test-server.kuma-test.svc.cluster.local",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
