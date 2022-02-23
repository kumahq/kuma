package reachableservices

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ReachableServicesOnK8s() {
	var cluster Cluster

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone)).
			Install(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin`)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(testserver.Install(
				testserver.WithName("client-server"),
				testserver.WithReachableServices("first-test-server_kuma-test_svc_80"),
			)).
			Install(testserver.Install(
				testserver.WithName("first-test-server"),
			)).
			Install(testserver.Install(
				testserver.WithName("second-test-server"),
			)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should be able to connect only to reachable services", func() {
		// given the client
		pods, err := k8s.ListPodsE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "client-server"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod := pods[0]

		// when tries to connect to a reachable service
		_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "client-server",
			"curl", "-v", "-m", "3", "--fail", "first-test-server_kuma-test_svc_80.mesh")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// when trying to connect to non-reachable services via Kuma DNS
		_, _, err = cluster.Exec(TestNamespace, clientPod.GetName(), "client-server",
			"curl", "-v", "second-test-server_kuma-test_svc_80.mesh")

		// then it fails because Kuma DP has no such DNS
		Expect(err).To(HaveOccurred())

		// when trying to connect to non-reachable service via Kubernetes DNS
		_, _, err = cluster.Exec(TestNamespace, clientPod.GetName(), "client-server",
			"curl", "-v", "second-test-server")

		// then it fails because we don't encrypt traffic to unknown destination in the mesh
		Expect(err).To(HaveOccurred())
	})
}
