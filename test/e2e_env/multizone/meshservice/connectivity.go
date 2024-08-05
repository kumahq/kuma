package meshservice

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Connectivity() {
	namespace := "msconnectivity"
	meshName := "msconnectivity"
	autoGenerateUniversalClusterName := "autogenerate-universal"

	var autoGenerateUniversalCluster *UniversalCluster

	var testServerPodNames []string
	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(YamlUniversal(fmt.Sprintf(`
type: HostnameGenerator
name: kube-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.svc.{{ .Zone }}.mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/managed-by: k8s-controller
        kuma.io/mesh: %s
        k8s.kuma.io/is-headless-service: "false"
`, meshName))).
			Install(YamlUniversal(`
type: HostnameGenerator
name: kube-not-my-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.svc.{{ .Zone }}.not-my-mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/managed-by: k8s-controller
        kuma.io/mesh: non-existent-mesh
        k8s.kuma.io/is-headless-service: "false"
`)).
			Install(YamlUniversal(fmt.Sprintf(`
type: HostnameGenerator
name: uni-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.svc.{{ .Zone }}.mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/mesh: "%s"
        kuma.io/managed-by: meshservice-generator
`, meshName))).
			Install(YamlUniversal(`
type: HostnameGenerator
name: uni-not-my-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.svc.{{ .Zone }}.not-my-mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/mesh: non-existent-mesh
        kuma.io/managed-by: meshservice-generator
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
  kuma.io/managed-by: meshservice-generator # should be changed to kuma.io/env: universal long term
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
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server-1"}))).
			Install(YamlUniversal(uniServiceYAML)).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"}))).
			Install(YamlUniversal(uniServiceYAML)).
			Setup(multizone.UniZone2)
		Expect(err).ToNot(HaveOccurred())

		autoGenerateUniversalCluster = NewUniversalCluster(NewTestingT(), autoGenerateUniversalClusterName, Silent)
		err = NewClusterSetup().
			Install(Kuma(
				core.Zone,
				WithGlobalAddress(multizone.Global.GetKuma().GetKDSServerAddress()),
				WithEnv("KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY", "0s"), // we have only 1 Kuma CP instance so there is no risk setting this to 0
				WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
				WithEnv("KUMA_EXPERIMENTAL_GENERATE_MESH_SERVICES", "true"),
			)).
			Install(IngressUniversal(multizone.Global.GetKuma().GenerateZoneIngressToken)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "auto-uni-test-server"}))).
			Setup(autoGenerateUniversalCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugUniversal(multizone.UniZone2, meshName)
		DebugUniversal(autoGenerateUniversalCluster, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(autoGenerateUniversalCluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
		Expect(autoGenerateUniversalCluster.DismissCluster()).To(Succeed())
	})

	type testCase struct {
		address          func() string
		expectedInstance string
		should           types.GomegaMatcher
	}

	DescribeTable("client from Kubernetes",
		func(given testCase) {
			if given.should == nil {
				given.should = Succeed()
			}
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(multizone.KubeZone1, "demo-client", given.address(),
					client.FromKubernetesPod(meshName, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal(given.expectedInstance))
			}, "30s", "1s").Should(given.should)
		},
		Entry("should access service in the same Kubernetes cluster", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.cluster.local:80" },
			expectedInstance: "kube-test-server-1",
		}),
		Entry("should access service in the same Kubernetes cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.kuma-1.mesh-specific.mesh.local:80" },
			expectedInstance: "kube-test-server-1",
		}),
		Entry("should access headless service in the same Kubernetes cluster", testCase{
			address: func() string {
				return fmt.Sprintf("http://%s.statefulset-test-server.msconnectivity.svc.cluster.local:80", testServerPodNames[0])
			},
			expectedInstance: "kube-statefulset-test-server-1",
		}),
		Entry("should access service in another Kubernetes cluster", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.kuma-2.mesh.local:80" },
			expectedInstance: "kube-test-server-2",
		}),
		Entry("should access service in another Universal cluster", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should access service in another Universal cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should not be able to access service in another Universal cluster if the hostname generator for that name doesn't match", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.not-my-mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server",
			should:           Not(Succeed()),
		}),
		Entry("should access service in another Universal cluster where MeshService is autogenerated", testCase{
			address:          func() string { return "http://test-server.svc.autogenerate-universal.mesh.local:80" },
			expectedInstance: "auto-uni-test-server",
		}),
	)

	DescribeTable("client from Universal",
		func(given testCase) {
			if given.should == nil {
				given.should = Succeed()
			}
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(multizone.UniZone1, "uni-demo-client", given.address())
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal(given.expectedInstance))
			}, "30s", "1s").Should(given.should)
		},
		Entry("should access the headless service in another Kubernetes cluster 1", testCase{
			address: func() string {
				return fmt.Sprintf("http://%s.statefulset-test-server.msconnectivity.svc.kuma-1.mesh.local:80", testServerPodNames[0])
			},
			expectedInstance: "kube-statefulset-test-server-1",
		}),
		Entry("should access service in another Kubernetes cluster 2", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.kuma-2.mesh.local:80" },
			expectedInstance: "kube-test-server-2",
		}),
		Entry("should access service in a Kubernetes cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.kuma-2.mesh-specific.mesh.local:80" },
			expectedInstance: "kube-test-server-2",
		}),
		Entry("should not be able to access service in a Kubernetes cluster if the hostname generator for that name doesn't match", testCase{
			address: func() string {
				return "http://test-server.msconnectivity.svc.kuma-2.not-my-mesh-specific.mesh.local:80"
			},
			expectedInstance: "kube-test-server-2",
			should:           Not(Succeed()),
		}),
		Entry("should access service in the same Universal cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.svc.kuma-4.mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server-1",
		}),
		Entry("should access service in another Universal cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should access service in another Universal cluster", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should access service in another Universal cluster where MeshService is autogenerated", testCase{
			address:          func() string { return "http://test-server.svc.autogenerate-universal.mesh.local:80" },
			expectedInstance: "auto-uni-test-server",
		}),
	)
}
