package meshservice

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Connectivity() {
	namespace := "msconnectivity"
	clientNamespace := "msconnectivity-client"
	meshName := "msconnectivity"
	autoGenerateUniversalClusterName := "autogenerate-universal"

	var autoGenerateUniversalCluster *UniversalCluster

	var testServerPodNames []string
	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(samples.MeshMTLSBuilder().
				WithName(meshName).
				WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).
				WithPermissiveMTLSBackends(),
			)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(YamlUniversal(fmt.Sprintf(`
type: HostnameGenerator
name: kube-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.svc.{{ .Zone }}.mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: kubernetes
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
        kuma.io/env: kubernetes
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
        kuma.io/env: universal
        kuma.io/mesh: "%s"
`, meshName))).
			Install(YamlUniversal(`
type: HostnameGenerator
name: uni-not-my-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.svc.{{ .Zone }}.not-my-mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: universal
        kuma.io/mesh: non-existent-mesh
`)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		meshGateway := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: edge-gateway-ms
  labels:
    kuma.io/origin: zone
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: edge-gateway-ms_%s_svc
  conf:
    listeners:
    - port: 8080
      protocol: HTTP
`, meshName, namespace)
		gatewayRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: route-ms
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway-ms
  to:
    - targetRef:
        kind: Mesh
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /local
          default:
            backendRefs:
              - kind: MeshService
                name: test-server
                namespace: %s
                port: 80
                weight: 1
        - matches:
            - path:
                type: PathPrefix
                value: /uni-2
          default:
            backendRefs:
              - kind: MeshService
                labels:
                  kuma.io/display-name: test-server
                  kuma.io/zone: kuma-5
                port: 80
                weight: 1
        - matches:
            - path:
                type: PathPrefix
                value: /no-matching-backend
          default:
            backendRefs:
              - kind: MeshService
                labels:
                  kuma.io/display-name: this-doesnt-exist
                port: 80
                weight: 1
`, Config.KumaNamespace, meshName, namespace)
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(clientNamespace)).
			Install(testserver.Install(
				testserver.WithName("demo-client"),
				testserver.WithNamespace(clientNamespace),
			)).
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
			Install(YamlK8s(meshGateway)).
			Install(YamlK8s(gatewayRoute)).
			Install(YamlK8s(gateway.MkGatewayInstance("edge-gateway-ms", namespace, meshName))).
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

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
			)).
			Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(DemoClientUniversal("uni-demo-client", meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server-1"}))).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"}))).
			Setup(multizone.UniZone2)
		Expect(err).ToNot(HaveOccurred())

		autoGenerateUniversalCluster = NewUniversalCluster(NewTestingT(), autoGenerateUniversalClusterName, Silent)
		err = NewClusterSetup().
			Install(Kuma(
				core.Zone,
				WithGlobalAddress(multizone.Global.GetKuma().GetKDSServerAddress()),
				WithEnv("KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY", "0s"), // we have only 1 Kuma CP instance so there is no risk setting this to 0
				WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
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
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
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

	DescribeTable("Gateway in Kubernetes",
		func(given testCase) {
			if given.should == nil {
				given.should = Succeed()
			}
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					multizone.KubeZone1, "demo-client",
					fmt.Sprintf("http://edge-gateway-ms.%s:8080/%s", namespace, given.address()),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal(given.expectedInstance))
			}, "30s", "1s").Should(Succeed())
		},
		Entry("should access service in the same Kubernetes cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "local" },
			expectedInstance: "kube-test-server-1",
		}),
		Entry("should access service in the a Universal cluster", testCase{
			address:          func() string { return "uni-2" },
			expectedInstance: "uni-test-server",
		}),
	)

	type failureCase struct {
		path         string
		responseCode int
	}
	DescribeTable("Gateway in Kubernetes with incorrect routes",
		func(given failureCase) {
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					multizone.KubeZone1, "demo-client",
					fmt.Sprintf("http://edge-gateway-ms.%s:8080/%s", namespace, given.path),
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(given.responseCode))
			}, "30s", "1s").Should(Succeed())
		},
		Entry("no matching backend should return 500", failureCase{
			path:         "no-matching-backend",
			responseCode: 500,
		}),
	)

	DescribeTable("client from Kubernetes",
		func(given testCase) {
			if given.should == nil {
				given.should = Succeed()
			}
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(multizone.KubeZone1, "demo-client", given.address(),
					client.FromKubernetesPod(namespace, "demo-client"),
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
		Entry("should access service in a Universal cluster", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should access service in a Universal cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should not be able to access service in a Universal cluster if the hostname generator for that name doesn't match", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.not-my-mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server",
			should:           Not(Succeed()),
		}),
		Entry("should access service in a Universal cluster where MeshService is autogenerated", testCase{
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
