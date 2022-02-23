package helm

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func AppDeploymentWithHelmChart() {
	defaultMesh := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
`

	var cluster Cluster

	BeforeEach(func() {
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
				WithCPReplicas(3),         // test HA capability
				WithCNI(),
			)).
			Install(YamlK8s(defaultMesh)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default")).
			Install(testserver.Install()).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		// tear down apps
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		// tear down Kuma
		Expect(cluster.DeleteKuma()).To(Succeed())
		// tear down cluster
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("Should deploy two apps", func() {
		pods, err := k8s.ListPodsE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod := pods[0]

		Eventually(func() (string, error) {
			_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server")
			return stderr, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

		Eventually(func() (string, error) {
			_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.mesh")
			return stderr, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

		Eventually(func() (string, error) { // should access a service with . instead of _
			_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server.kuma-test.svc.80.mesh")
			return stderr, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))
	})
}
