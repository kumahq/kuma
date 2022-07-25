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

	var setup = func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithSkipDefaultMesh(true), // it's common case for HELM deployments that Mesh is also managed by HELM therefore it's not created by default
				WithHelmOpt("cni.experimental.sleepBeforeRunSeconds", "60"),
				WithExperimentalCNI(),
			)).
			Install(YamlK8s(defaultMesh)).
			Setup(cluster)
		// here we could patch the "command" of the CNI, kubectl patch ...
		Expect(err).ToNot(HaveOccurred())
	}

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
		Expect(cluster.DeleteNode("k3d-second-node-0")).To(Succeed())
	})

	It(
		"Should deploy two apps",
		func() {
			setup()

			// write a test case that shows the test-server does not come up cleanly without taint-controller

			err := cluster.CreateNode("second-node", "second=true")
			Expect(err).ToNot(HaveOccurred())

			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(TestNamespace)).
				Install(DemoClientK8sWithAffinity("default", TestNamespace)).
				Install(testserver.Install(func(opts *testserver.DeploymentOpts) {
					opts.NodeSelector = map[string]string{
						"second": "true",
					}
				})).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())
			// assert pods demo-client and testserver are available on the node

			clientPodName, err := PodNameOfApp(cluster, "demo-client", TestNamespace)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() (string, error) {
				_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server")
				return stderr, err
			}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

			Eventually(func() (string, error) {
				_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.mesh")
				return stderr, err
			}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

			Eventually(func() (string, error) { // should access a service with . instead of _
				_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server.kuma-test.svc.80.mesh")
				return stderr, err
			}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))
		},
	)
}
