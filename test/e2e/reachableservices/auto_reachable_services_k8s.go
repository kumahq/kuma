package reachableservices

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	admintunnel "github.com/kumahq/kuma/test/framework/envoy_admin/tunnel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var k8sCluster Cluster

var _ = E2EBeforeSuite(func() {
	k8sCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

	err := NewClusterSetup().
		Install(Kuma(config_core.Standalone,
			WithEnv("KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES", "true"),
		)).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(testserver.Install(testserver.WithName("client-server"), testserver.WithMesh("default"))).
		Install(testserver.Install(testserver.WithName("first-test-server"), testserver.WithMesh("default"))).
		Install(testserver.Install(testserver.WithName("second-test-server"), testserver.WithMesh("default"))).
		Setup(k8sCluster)

	Expect(err).ToNot(HaveOccurred())

	E2EDeferCleanup(func() {
		Expect(k8sCluster.DeleteKuma()).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})
})

func AutoReachableServices() {
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

	It("should not connect to non auto reachable service", func() {
		// given
		removeDefaultTrafficPermission()
		noDefaultTrafficPermission()

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp1
  namespace: %s
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshService
    name: first-test-server_kuma-test_svc_80
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Deny
    - targetRef:
        kind: MeshService
        name: client-server_kuma-test_svc_80
      default:
        action: Allow
`, Config.KumaNamespace))(k8sCluster)).To(Succeed())

		// then
		Eventually(func() int {
			return numberOfClusterConfiguration("second-test-server", "first-test-server_kuma-test_svc_80")
		}).Should(Equal(0))

		Eventually(func() int {
			return numberOfClusterConfiguration("client-server", "first-test-server_kuma-test_svc_80")
		}, "30s", "1s").Should(Equal(1))
	})
}

func numberOfClusterConfiguration(appName string, targetCluster string) int {
	pod, err := PodNameOfApp(k8sCluster, appName, TestNamespace)
	Expect(err).ToNot(HaveOccurred())
	tunnel := k8s.NewTunnel(k8sCluster.GetKubectlOptions(TestNamespace), k8s.ResourceTypePod, pod, 0, 9901)
	tunnel.ForwardPort(k8sCluster.GetTesting())
	admin := admintunnel.NewK8sEnvoyAdminTunnel(k8sCluster.GetTesting(), tunnel.Endpoint())

	// then
	// first-test-server_kuma-test_svc_80::added_via_api::true
	stat, err := admin.GetStats(fmt.Sprintf("%s::added_via_api", targetCluster))
	Expect(err).ToNot(HaveOccurred())
	return len(stat.Stats)
}
