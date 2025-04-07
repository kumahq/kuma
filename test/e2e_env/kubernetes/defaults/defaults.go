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

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName)
	})

	AfterAll(func() {
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	policyCreated := func(typ, name string, namespace ...string) func() bool {
		return func() bool {
			output, err := k8s.RunKubectlAndGetOutputE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace...), "get", typ)
			if err != nil {
				return false
			}
			return strings.Contains(output, name)
		}
	}

	It("should create default policies for default mesh", func() {
		Eventually(policyCreated("trafficpermission", "allow-all-default"), "30s", "1s").MustPassRepeatedly(3).Should(BeFalse())
		Eventually(policyCreated("trafficroute", "route-all-default"), "30s", "1s").MustPassRepeatedly(3).Should(BeFalse())
		Eventually(policyCreated("meshtimeout", "mesh-timeout-all-outbounds-default", Config.KumaNamespace), "30s", "1s").MustPassRepeatedly(3).Should(BeTrue())
		Eventually(policyCreated("meshtimeout", "mesh-timeout-all-inbounds-default", Config.KumaNamespace), "30s", "1s").MustPassRepeatedly(3).Should(BeTrue())
		Eventually(policyCreated("meshcircuitbreaker", "mesh-circuit-breaker-all-default", Config.KumaNamespace), "30s", "1s").MustPassRepeatedly(3).Should(BeTrue())
		Eventually(policyCreated("meshretry", "mesh-retry-all-default", Config.KumaNamespace), "30s", "1s").MustPassRepeatedly(3).Should(BeTrue())
	})

	It("should create default policies for non-default mesh", func() {
		Eventually(policyCreated("trafficpermission", "allow-all-"+meshName), "30s", "1s").Should(BeFalse())
		Eventually(policyCreated("trafficroute", "route-all-"+meshName), "30s", "1s").Should(BeFalse())
		Eventually(policyCreated("meshtimeout", "mesh-timeout-all-outbounds-"+meshName, Config.KumaNamespace), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("meshtimeout", "mesh-timeout-all-inbounds-"+meshName, Config.KumaNamespace), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("meshcircuitbreaker", "mesh-circuit-breaker-all-"+meshName, Config.KumaNamespace), "30s", "1s").Should(BeTrue())
		Eventually(policyCreated("meshretry", "mesh-retry-all-"+meshName, Config.KumaNamespace), "30s", "1s").Should(BeTrue())
	})

	It("should create a zone", func() {
		Eventually(policyCreated("zone", "default"), "30s", "1s").Should(BeTrue())
	})
}
