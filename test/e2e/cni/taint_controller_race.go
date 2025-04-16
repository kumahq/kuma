package cni

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func AppDeploymentWithCniAndTaintController() {
	defaultMesh := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
`

	var cluster Cluster
	var k8sCluster *K8sCluster
	nodeName := fmt.Sprintf(
		"second-%s",
		strings.ToLower(random.UniqueId()),
	)

	setup := func() {
		k8sCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		cluster = k8sCluster.
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		err := NewClusterSetup().
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithSkipDefaultMesh(true), // it's common case for HELM deployments that Mesh is also managed by HELM therefore it's not created by default
				WithHelmOpt("cni.delayStartupSeconds", "40"),
				WithCNI(),
			)).
			Install(YamlK8s(defaultMesh)).
			Setup(cluster)
		// here we could patch the "command" of the CNI, kubectl patch ...
		Expect(err).ToNot(HaveOccurred())
	}

	AfterEachFailure(func() {
		DebugKube(k8sCluster, "default", TestNamespace)
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
		Expect(k8sCluster.DeleteNode("k3d-" + nodeName + "-0")).To(Succeed())
	})

	It(
		"prevents race condition following a slow CNI start",
		func() {
			setup()

			err := k8sCluster.CreateNode(nodeName, "second=true")
			Expect(err).ToNot(HaveOccurred())

			Expect(k8sCluster.LoadImages(
				Config.KumaDPImageRepo,
				Config.KumaCNIImageRepo,
				Config.KumaInitImageRepo,
				Config.KumaUniversalImageRepo,
			)).ToNot(HaveOccurred())

			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(TestNamespace)).
				Install(testserver.Install(testserver.WithNodeSelector(map[string]string{"second": "true"}))).
				Install(democlient.Install(democlient.WithNamespace(TestNamespace), democlient.WithNodeSelector(map[string]string{"second": "true"}))).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					cluster, "demo-client", "test-server",
					client.FromKubernetesPod(TestNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "10s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					cluster, "demo-client", "test-server_kuma-test_svc_80.mesh",
					client.FromKubernetesPod(TestNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "10s", "1s").Should(Succeed())

			Eventually(func(g Gomega) { // should access a service with . instead of _
				_, err := client.CollectEchoResponse(
					cluster, "demo-client", "test-server.kuma-test.svc.80.mesh",
					client.FromKubernetesPod(TestNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "10s", "1s").Should(Succeed())
		},
	)
}
