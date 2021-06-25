package helm

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

func AppDeploymentWithHelmChart() {
	namespaceWithSidecarInjection := func(namespace string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  annotations:
    kuma.io/sidecar-injection: "enabled"
`, namespace)
	}

	defaultMesh := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
`

	var cluster Cluster
	var deployOptsFuncs = KumaK8sDeployOpts

	BeforeEach(func() {
		c, err := NewK8sClusterWithTimeout(
			NewTestingT(),
			Kuma1,
			Silent,
			6*time.Second)
		Expect(err).ToNot(HaveOccurred())

		cluster = c.WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)
		deployOptsFuncs = append(deployOptsFuncs,
			WithInstallationMode(HelmInstallationMode),
			WithHelmReleaseName(releaseName),
			WithSkipDefaultMesh(true), // it's common case for HELM deployments that Mesh is also managed by HELM therefore it's not created by default
			WithCPReplicas(3),         // test HA capability
			WithCNI())

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Install(YamlK8s(defaultMesh)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s("default")).
			Install(EchoServerK8s("default")).
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
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
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
				"curl", "-v", "-m", "3", "--fail", "echo-server")
			return stderr, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

		Eventually(func() (string, error) {
			_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
				"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_80.mesh")
			return stderr, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

		Eventually(func() (string, error) { // should access a service with . instead of _
			_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
				"curl", "-v", "-m", "3", "--fail", "echo-server.kuma-test.svc.80.mesh")
			return stderr, err
		}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))
	})
}
