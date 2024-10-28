package reachableservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func AutoReachableServices() {
	esNamespace := "external-service"
	namespace := "reachable-services"
	meshName := "reachable-services"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(esNamespace)).
			Install(MTLSMeshKubernetes(meshName)).
			Install(testserver.Install(testserver.WithName("client-server"), testserver.WithMesh(meshName), testserver.WithNamespace(namespace))).
			Install(testserver.Install(testserver.WithName("first-test-server"), testserver.WithMesh(meshName), testserver.WithNamespace(namespace))).
			Install(testserver.Install(testserver.WithName("second-test-server"), testserver.WithMesh(meshName), testserver.WithNamespace(namespace))).
			Install(testserver.Install(
				testserver.WithName("external-http-service"),
				testserver.WithNamespace(esNamespace),
				testserver.WithEchoArgs("echo", "--instance", "external-http-service"),
			)).
			Setup(KubeCluster)

		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(KubeCluster, meshName, v1alpha1.MeshTrafficPermissionResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(KubeCluster.DeleteNamespace(namespace)).To(Succeed())
		Expect(KubeCluster.DeleteNamespace(esNamespace)).To(Succeed())
		Expect(KubeCluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should not connect to non auto reachable service", func() {
		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: rs-client-to-server
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: first-test-server_reachable-services_svc_80
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Deny
    - targetRef:
        kind: MeshService
        name: client-server_reachable-services_svc_80
      default:
        action: Allow
`, Config.KumaNamespace, meshName))(KubeCluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(KubeCluster, "second-test-server", namespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := KubeCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod+"."+namespace, "--mesh", meshName, "--type=clusters")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Not(ContainSubstring("first-test-server_reachable-services_svc_80")))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(KubeCluster, "client-server", namespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := KubeCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod+"."+namespace, "--mesh", meshName, "--type=clusters")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("first-test-server_reachable-services_svc_80"))
		}, "30s", "1s").Should(Succeed())

		Consistently(func(g Gomega) {
			failures, err := client.CollectFailure(
				KubeCluster,
				"second-test-server",
				"first-test-server:80",
				client.FromKubernetesPod(namespace, "second-test-server"),
			)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(failures.Exitcode).To(Equal(52))
		}, "5s", "1s").Should(Succeed())
	})

	It("should connect to ExternalService when MeshTrafficPermission defined", func() {
		// when
		Expect(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: reachable-services
metadata:
  name: external-service-1
spec:
  tags:
    kuma.io/service: external-http-service
    kuma.io/protocol: http
  networking:
    address: external-http-service.external-service.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
`)(KubeCluster)).To(Succeed())

		// then
		Consistently(func(g Gomega) {
			failures, err := client.CollectFailure(
				KubeCluster,
				"client-server",
				"external-http-service.mesh",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(failures.Exitcode).To(Equal(6))
		}, "5s", "1s").Should(Succeed())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: rs-mtp-es
  namespace: %s
  labels:
    kuma.io/mesh: reachable-services
spec:
  targetRef:
    kind: MeshService
    name: external-http-service
  from:
    - targetRef:
        kind: MeshService
        name: client-server_reachable-services_svc_80
      default:
        action: Allow
`, Config.KumaNamespace))(KubeCluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				KubeCluster,
				"client-server",
				"external-http-service.mesh",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(resp.Instance).To(Equal("external-http-service"))
		}, "30s", "1s").Should(Succeed())
	})
}
