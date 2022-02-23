package kubernetes

import (
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var k8sCluster Cluster

var _ = E2EBeforeSuite(func() {
	k8sClusters, err := NewK8sClusters([]string{Kuma1}, Silent)
	Expect(err).ToNot(HaveOccurred())

	k8sCluster = k8sClusters.GetCluster(Kuma1)

	Expect(Kuma(config_core.Standalone)(k8sCluster)).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(k8sCluster.DeleteKuma()).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})
})

func TrafficPermission() {
	removeDefaultTrafficPermission := func() {
		err := k8s.RunKubectlE(k8sCluster.GetTesting(), k8sCluster.GetKubectlOptions(), "delete", "trafficpermission", "allow-all-default")
		Expect(err).ToNot(HaveOccurred())
	}

	noDefaultTrafficPermission := func() {
		Eventually(func() bool {
			out, err := k8s.RunKubectlAndGetOutputE(k8sCluster.GetTesting(), k8sCluster.GetKubectlOptions(), "get", "trafficpermissions")
			if err != nil {
				return false
			}
			return !strings.Contains(out, "allow-all-default")
		}, "30s", "1s").Should(BeTrue())
	}

	defaultPoliciesCreated := func() {
		Eventually(func() bool {
			out, err := k8s.RunKubectlAndGetOutputE(k8sCluster.GetTesting(), k8sCluster.GetKubectlOptions(), "get", "meshes", "-o", "yaml")
			if err != nil {
				return false
			}
			return strings.Contains(out, "k8s.kuma.io/mesh-defaults-generated")
		}, "30s", "1s").Should(BeTrue())
	}

	restartKumaCP := func() {
		pods := k8sCluster.GetKuma().(*K8sControlPlane).GetKumaCPPods()
		Expect(pods).To(HaveLen(1))
		err := k8s.RunKubectlE(k8sCluster.GetTesting(), k8sCluster.GetKubectlOptions(), "delete", "pod", pods[0].GetName(), "-n", pods[0].GetNamespace())
		Expect(err).ToNot(HaveOccurred())
		Expect(k8sCluster.(*K8sCluster).WaitApp(Config.KumaServiceName, Config.KumaNamespace, 1)).To(Succeed())
	}

	It("should not create deleted default traffic permission after Kuma CP restart", func() {
		// given
		defaultPoliciesCreated()

		// when
		removeDefaultTrafficPermission()
		// then
		noDefaultTrafficPermission()

		// when
		restartKumaCP()
		// and when
		time.Sleep(10 * time.Second)
		// then
		noDefaultTrafficPermission()
	})
}
