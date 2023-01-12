package defaults

import (
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Defaults() {
	meshName := "defaults-non-default"

	BeforeAll(func() {
		Expect(kubernetes.Cluster.Install(MeshKubernetes(meshName))).To(Succeed())
	})

	AfterAll(func() {
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	policyCreated := func(typ, name string) func() bool {
		return func() bool {
			output, err := k8s.RunKubectlAndGetOutputE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(), "get", typ)
			if err != nil {
				return false
			}
			return strings.Contains(output, name)
		}
	}

	It("should create default policies for default mesh", func() {
		Eventually(policyCreated("trafficpermission", "allow-all-default"), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("trafficroute", "route-all-default"), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("timeout", "timeout-all-default"), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("circuitbreaker", "circuit-breaker-all-default"), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("retry", "retry-all-default"), "30s", "1s").Should(BeTrue())
	})

	It("should create default policies for non-default mesh", func() {
		Eventually(policyCreated("trafficpermission", "allow-all-"+meshName), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("trafficroute", "route-all-"+meshName), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("timeout", "timeout-all-"+meshName), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("circuitbreaker", "circuit-breaker-all-"+meshName), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("retry", "retry-all-"+meshName), "30s", "1s").Should(BeTrue())
	})
}
