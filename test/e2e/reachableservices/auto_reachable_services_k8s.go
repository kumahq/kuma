package reachableservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

var k8sCluster Cluster

var _ = E2EBeforeSuite(func() {
	esNamespace := "external-service"
	k8sCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

	err := NewClusterSetup().
		Install(Kuma(config_core.Zone,
			WithEnv("KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES", "true"),
		)).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(Namespace(esNamespace)).
		Install(MTLSMeshKubernetes("default")).
		Install(testserver.Install(testserver.WithName("client-server"), testserver.WithMesh("default"))).
		Install(testserver.Install(testserver.WithName("first-test-server"), testserver.WithMesh("default"))).
		Install(testserver.Install(testserver.WithName("second-test-server"), testserver.WithMesh("default"))).
		Install(testserver.Install(
			testserver.WithName("external-http-service"),
			testserver.WithNamespace(esNamespace),
			testserver.WithEchoArgs("echo", "--instance", "external-http-service"),
		)).
		Setup(k8sCluster)

	Expect(err).ToNot(HaveOccurred())

	E2EDeferCleanup(func() {
		Expect(k8sCluster.DeleteNamespace(esNamespace)).To(Succeed())
		Expect(k8sCluster.DeleteKuma()).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})
})

func AutoReachableServices() {
	E2EAfterEach(func() {
		Expect(DeleteMeshResources(k8sCluster, "default", v1alpha1.MeshTrafficPermissionResourceTypeDescriptor)).To(Succeed())
	})

	It("should not connect to non auto reachable service", func() {
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
		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(k8sCluster, "second-test-server", TestNamespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := k8sCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod+"."+TestNamespace, "--type=clusters")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Not(ContainSubstring("first-test-server_kuma-test_svc_80")))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(k8sCluster, "client-server", TestNamespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := k8sCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod+"."+TestNamespace, "--type=clusters")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("first-test-server_kuma-test_svc_80"))
		}, "30s", "1s").Should(Succeed())

		Consistently(func(g Gomega) {
			failures, err := client.CollectFailure(
				k8sCluster,
				"second-test-server",
				"first-test-server:80",
				client.FromKubernetesPod(TestNamespace, "second-test-server"),
			)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(failures.Exitcode).To(Equal(52))
		}, "30s", "1s").Should(Succeed())
	})

	It("should connect to ExternalService when MeshTrafficPermission defined", func() {
		// when
		Expect(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: default
metadata:
  name: external-service-1
spec:
  tags:
    kuma.io/service: external-http-service
    kuma.io/protocol: http
  networking:
    address: external-http-service.external-service.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
`)(k8sCluster)).To(Succeed())

		// then
		Consistently(func(g Gomega) {
			failures, err := client.CollectFailure(
				k8sCluster,
				"client-server",
				"external-http-service.mesh",
				client.FromKubernetesPod(TestNamespace, "client-server"),
			)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(failures.Exitcode).To(Or(Equal(6), Equal(28)))
		}, "30s", "1s").Should(Succeed())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp-es
  namespace: %s
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshService
    name: external-http-service
  from:
    - targetRef:
        kind: MeshService
        name: client-server_kuma-test_svc_80
      default:
        action: Allow
`, Config.KumaNamespace))(k8sCluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				k8sCluster,
				"client-server",
				"external-http-service.mesh",
				client.FromKubernetesPod(TestNamespace, "client-server"),
			)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(resp.Instance).To(Equal("external-http-service"))
		}, "30s", "1s").Should(Succeed())
	})
}
