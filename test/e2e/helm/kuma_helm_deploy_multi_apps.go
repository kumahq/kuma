package helm

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

func AppDeploymentWithHelmChart() {
	defaultMesh := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
`

	var cluster Cluster

	var setup = func(withCni KumaDeploymentOption) {
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
				withCni,
			)).
			Install(YamlK8s(defaultMesh)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default", TestNamespace)).
			Install(testserver.Install()).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	}

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	DescribeTable(
		"Should deploy two apps",
		func(withCni KumaDeploymentOption) {
			setup(withCni)

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
		Entry("with default cni", WithCNI()),
		Entry("with new cni (experimental)", WithExperimentalCNI()),
	)
}
