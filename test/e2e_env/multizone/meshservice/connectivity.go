package meshservice

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Connectivity() {
	namespace := "msconnectivity"
	meshName := "msconnectivity"

	var testServerPodNames []string
	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(YamlUniversal(`
type: HostnameGenerator
name: kube-msconnectivity
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.{{ .Zone }}.k8s.msconnectivity'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: global
        kuma.io/managed-by: k8s-controller
        k8s.kuma.io/is-headless-service: "false"
`)).
			Install(YamlUniversal(`
type: HostnameGenerator
name: kube-msconnectivity-headless
spec:
  template: '{{ label "statefulset.kubernetes.io/pod-name" }}.{{ label "k8s.kuma.io/service-name" }}.{{ .Namespace }}.{{ .Zone }}.k8s.msconnectivity'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: global
        kuma.io/managed-by: k8s-controller
        k8s.kuma.io/is-headless-service: "true"
`)).
			Install(YamlUniversal(`
type: HostnameGenerator
name: uni-msconnectivity
spec:
  template: '{{ .DisplayName }}.{{ .Zone }}.universal.msconnectivity'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: global
        kuma.io/env: universal
`)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-1"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithName("statefulset-test-server"),
				testserver.WithStatefulSet(),
				testserver.WithHeadlessService(),
				testserver.WithEchoArgs("echo", "--instance", "kube-statefulset-test-server-1"),
			)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		Expect(multizone.KubeZone1.WaitApp("statefulset-test-server", namespace, 1)).To(Succeed())

		for _, pod := range k8s.ListPods(multizone.KubeZone1.GetTesting(),
			multizone.KubeZone1.GetKubectlOptions(namespace),
			kube_meta.ListOptions{
				LabelSelector: "app=statefulset-test-server",
			},
		) {
			testServerPodNames = append(testServerPodNames, pod.Name)
		}
		Expect(testServerPodNames).To(HaveLen(1))

		kubeServiceYAML := `
apiVersion: kuma.io/v1alpha1
kind: MeshService
metadata:
  name: test-server
  namespace: msconnectivity
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: msconnectivity
    kuma.io/managed-by: k8s-controller
    k8s.kuma.io/is-headless-service: "false"
spec:
  selector:
    dataplaneTags:
      app: test-server
      k8s.kuma.io/namespace: msconnectivity
  ports:
  - port: 80
    name: main
    targetPort: main
    appProtocol: http
`
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
			)).
			Install(YamlK8s(kubeServiceYAML)).
			Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())

		uniServiceYAML := `
type: MeshService
name: test-server
mesh: msconnectivity
labels:
  kuma.io/origin: zone
  kuma.io/env: universal
spec:
  selector:
    dataplaneTags:
      kuma.io/service: test-server
  ports:
  - port: 80
    targetPort: 80
    appProtocol: http
`

		err = NewClusterSetup().
			Install(DemoClientUniversal("uni-demo-client", meshName, WithTransparentProxy(true))).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"}))).
			Install(YamlUniversal(uniServiceYAML)).
			Setup(multizone.UniZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugUniversal(multizone.UniZone2, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	type testCase struct {
		address          func() string
		expectedInstance string
	}

	DescribeTable("client from Kubernetes",
		func(given testCase) {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(multizone.KubeZone1, "demo-client", given.address(),
					client.FromKubernetesPod(meshName, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal(given.expectedInstance))
			}, "30s", "1s").Should(Succeed())
		},
		Entry("should access service in the same Kubernetes cluster", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.cluster.local:80" },
			expectedInstance: "kube-test-server-1",
		}),
		Entry("should access headless service in the same Kubernetes cluster", testCase{
			address: func() string {
				return fmt.Sprintf("http://%s.statefulset-test-server.msconnectivity.svc.cluster.local:80", testServerPodNames[0])
			},
			expectedInstance: "kube-statefulset-test-server-1",
		}),
		Entry("should access service in another Kubernetes cluster", testCase{
			address:          func() string { return "http://test-server.msconnectivity.kuma-2.k8s.msconnectivity:80" },
			expectedInstance: "kube-test-server-2",
		}),
		Entry("should access service in another Universal cluster", testCase{
			address:          func() string { return "http://test-server.kuma-5.universal.msconnectivity:80" },
			expectedInstance: "uni-test-server",
		}),
	)

	DescribeTable("client from Universal",
		func(given testCase) {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(multizone.UniZone1, "uni-demo-client", given.address())
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal(given.expectedInstance))
			}, "30s", "1s").Should(Succeed())
		},
		Entry("should access the headless service in another Kubernetes cluster 1", testCase{
			address: func() string {
				return fmt.Sprintf("http://%s.statefulset-test-server.msconnectivity.kuma-1.k8s.msconnectivity:80", testServerPodNames[0])
			},
			expectedInstance: "kube-statefulset-test-server-1",
		}),
		Entry("should access service in another Kubernetes cluster 2", testCase{
			address:          func() string { return "http://test-server.msconnectivity.kuma-2.k8s.msconnectivity:80" },
			expectedInstance: "kube-test-server-2",
		}),
		Entry("should access service in another Universal cluster", testCase{
			address:          func() string { return "http://test-server.kuma-5.universal.msconnectivity:80" },
			expectedInstance: "uni-test-server",
		}),
	)
}
