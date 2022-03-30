package deploy

import (
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func KubernetesDeployment() {
	var k8sCluster Cluster
	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		k8sCluster = k8sClusters.GetCluster(Kuma1)
		Expect(Kuma(config_core.Standalone)(k8sCluster)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(k8sCluster.DeleteKuma()).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})

	policyCreated := func(typ, name string) func() bool {
		return func() bool {
			output, err := k8s.RunKubectlAndGetOutputE(k8sCluster.GetTesting(), k8sCluster.GetKubectlOptions("default"), "get", typ)
			if err != nil {
				return false
			}
			return strings.Contains(output, name)
		}
	}

	It("should create default policies for default mesh", func() {
		Eventually(policyCreated("trafficpermission", "allow-all-default"), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("trafficroute", "route-all-default"), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("timeout", "timeout-all-default"), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("circuitbreaker", "circuit-breaker-all-default"), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("retry", "retry-all-default"), "5s", "500ms").Should(BeTrue())
	})

	It("should create default policies for non-default mesh", func() {
		Expect(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: non-default`)(k8sCluster)).To(Succeed())

		Eventually(policyCreated("trafficpermission", "allow-all-non-default"), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("trafficroute", "route-all-non-default"), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("timeout", "timeout-all-non-default"), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("circuitbreaker", "circuit-breaker-all-non-default"), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("retry", "retry-all-non-default"), "5s", "500ms").Should(BeTrue())
	})
}
