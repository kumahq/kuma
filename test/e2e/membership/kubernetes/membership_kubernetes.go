package kubernetes

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func MembershipKubernetes() {
	var cluster Cluster

	const kumaTestDefaultNs = "kuma-test-default"
	const kumaTestDemoNs = "kuma-test-demo"

	meshAllowingNamespace := func(mesh, ns string) string {
		return fmt.Sprintf(`apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  constraints:
    dataplaneProxy:
      requirements:
      - tags:
          k8s.kuma.io/namespace: %s`, mesh, ns)
	}

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(NamespaceWithSidecarInjection(kumaTestDefaultNs)).
			Install(NamespaceWithSidecarInjection(kumaTestDemoNs)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(kumaTestDefaultNs)).To(Succeed())
		Expect(cluster.DeleteNamespace(kumaTestDemoNs)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should take into account membership when dp is connecting to the CP", func() {
		// given
		Expect(YamlK8s(meshAllowingNamespace("default", kumaTestDefaultNs))(cluster)).To(Succeed())
		Expect(YamlK8s(meshAllowingNamespace("demo", kumaTestDemoNs))(cluster)).To(Succeed())

		// when
		err := testserver.Install(
			testserver.WithNamespace(kumaTestDefaultNs),
			testserver.WithMesh("default"),
		)(cluster)

		// then
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() (string, error) {
			return cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
		}, "30s", "1s").Should(ContainSubstring("test-server"))

		// when trying to change mesh to demo
		err = testserver.Install(
			testserver.WithNamespace(kumaTestDefaultNs),
			testserver.WithMesh("demo"),
			testserver.WithoutWaitingToBeReady(),
		)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then the client is not allowed to do it
		Eventually(func() (string, error) {
			return cluster.GetKuma().GetKumaCPLogs()
		}, "30s", "1s").Should(ContainSubstring("dataplane cannot be a member of mesh"))
		out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "--mesh", "demo")
		Expect(err).ToNot(HaveOccurred())
		Expect(out).ToNot(ContainSubstring("test-server"))

		// when a new client is deployed in demo namespace in demo mesh
		err = testserver.Install(
			testserver.WithNamespace(kumaTestDemoNs),
			testserver.WithMesh("demo"),
		)(cluster)

		// then it's allowed
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() (string, error) {
			return cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes", "--mesh", "demo")
		}, "30s", "1s").Should(ContainSubstring("test-server"))
	})
}
