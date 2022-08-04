package defaults

import (
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Defaults() {
	meshName := "defaults-non-default"

	BeforeAll(func() {
		Expect(env.Cluster.Install(MeshKubernetes(meshName))).To(Succeed())
	})

	AfterAll(func() {
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	policyCreated := func(typ, name string) func() bool {
		return func() bool {
			output, err := k8s.RunKubectlAndGetOutputE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(), "get", typ)
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
		Eventually(policyCreated("trafficpermission", "allow-all-"+meshName), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("trafficroute", "route-all-"+meshName), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("timeout", "timeout-all-"+meshName), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("circuitbreaker", "circuit-breaker-all-"+meshName), "5s", "500ms").Should(BeTrue())
		Eventually(policyCreated("retry", "retry-all-"+meshName), "5s", "500ms").Should(BeTrue())
	})
}
