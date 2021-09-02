package kubernetes

import (
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func VirtualProbes() {
	var k8sCluster Cluster
	var optsKubernetes = KumaK8sDeployOpts

	E2EBeforeSuite(func() {
		k8sClusters, err := NewK8sClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		k8sCluster = k8sClusters.GetCluster(Kuma1)

		Expect(Kuma(config_core.Standalone, optsKubernetes...)(k8sCluster)).To(Succeed())
		Expect(NamespaceWithSidecarInjection(TestNamespace)(k8sCluster)).To(Succeed())
		Expect(k8sCluster.VerifyKuma()).To(Succeed())
	})

	E2EAfterSuite(func() {
		Expect(k8sCluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(k8sCluster.DeleteKuma(optsKubernetes...)).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})

	testServerReady := func() bool {
		output, err := k8s.RunKubectlAndGetOutputE(k8sCluster.GetTesting(), k8sCluster.GetKubectlOptions(TestNamespace), "get", "pods")
		if err != nil {
			return false
		}
		lines := strings.Split(output, "\n")
		if len(lines) != 2 {
			return false
		}
		// 0:NAME 1:READY 2:STATUS 3:RESTARTS 4:AGE
		words := strings.Fields(lines[1])
		if len(words) != 5 {
			return false
		}
		if words[1] == "2/2" && words[2] == "Running" && words[3] == "0" {
			return true
		}
		return false
	}

	It("should deploy test-server with probes", func() {
		Expect(testserver.Install()(k8sCluster)).To(Succeed())

		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			Expect(testServerReady()).To(BeTrue())
		}
	})
}
