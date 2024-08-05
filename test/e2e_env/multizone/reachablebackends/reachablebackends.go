package reachablebackends

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func ReachableBackends() {
	meshName := "reachable-backends"
	namespace := "reachable-backends"
	namespaceOutside := "reachable-backends-non-mesh"
	reachableBackends := fmt.Sprintf(`
      refs:
      - kind: MeshService
        name: first-test-server
        namespace: %s
      - kind: MeshExternalService
        labels:
          kuma.io/access: external-service
      - kind: MeshMultiZoneService
        labels:
          reachable: "true"
`, namespace)

	meshPassthrough := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1 
kind: MeshPassthrough
metadata:
  name: disable-passthrough-reachable
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshSubset
    proxyTypes: ["Sidecar"]
    tags:
      kuma.io/service: client-server_reachable-backends_svc_80
  default:
    passthroughMode: None`, Config.KumaNamespace, meshName)

	meshExternalService := func(serviceName string) string {
		return fmt.Sprintf(`
type: MeshExternalService
name: %s-reachable
mesh: %s
labels:
  kuma.io/access: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s.reachable-backends-non-mesh.svc.cluster.local
      port: 80
`, serviceName, meshName, serviceName, serviceName)
	}

	hostnameGeneratorMs := fmt.Sprintf(`
type: HostnameGenerator
name: hg-ms-reachable
spec:
  selector:
    meshService:
      matchLabels:
        k8s.kuma.io/namespace: %s
  template: "{{ .DisplayName }}.mesh"
`, namespace)

	hostnameGeneratorMes := fmt.Sprintf(`
type: HostnameGenerator
name: hg-mes-reachable
spec:
  selector:
    meshExternalService:
      matchLabels:
        k8s.kuma.io/namespace: %s
  template: "{{ .DisplayName }}.mesh"
`, namespaceOutside)

	hostnameGeneratorMmzs := `
type: HostnameGenerator
name: hg-mmzs-reachable
spec:
  template: '{{ .DisplayName }}.global.mmzsreachable'
  selector:
    meshMultiZoneService:
      matchLabels:
        test-name: mmzsreachable
`

	mmzs := fmt.Sprintf(`
type: MeshMultiZoneService
name: other-zone-test-server
mesh: %s
labels:
  reachable: "true"
  test-name: mmzsreachable
spec:
  ports:
  - port: 80
    appProtocol: http
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: other-zone-test-server
        k8s.kuma.io/namespace: %s
`, meshName, namespace)

	mmzsNotAccessible := fmt.Sprintf(`
type: MeshMultiZoneService
name: other-zone-not-accessible
mesh: %s
labels:
  reachable: "false"
  test-name: mmzsreachable
spec:
  ports:
  - port: 80
    appProtocol: http
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: other-zone-not-accessible
        k8s.kuma.io/namespace: %s
`, meshName, namespace)

	BeforeAll(func() {
		// Global
		err := NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(YamlUniversal(hostnameGeneratorMes)).
			Install(YamlUniversal(hostnameGeneratorMs)).
			Install(YamlUniversal(hostnameGeneratorMmzs)).
			Install(YamlUniversal(mmzs)).
			Install(YamlUniversal(mmzsNotAccessible)).
			Install(YamlUniversal(meshExternalService("external-service"))).
			Install(YamlUniversal(meshExternalService("not-accessible-es"))).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		// Zone Kube1
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceOutside)).
			Install(YamlK8s(meshPassthrough)).
			Install(testserver.Install(
				testserver.WithName("client-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithReachableBackends(reachableBackends),
				testserver.WithReachableServices("non-existing"), // non existing so we don't get non targetRef outbounds
			)).
			Install(testserver.Install(
				testserver.WithName("client-server-no-access"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithReachableBackends("{}"),
				testserver.WithReachableServices("non-existing"), // non existing so we don't get non targetRef outbounds
			)).
			Install(testserver.Install(
				testserver.WithName("first-test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Install(testserver.Install(
				testserver.WithName("second-test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Install(testserver.Install(
				testserver.WithName("external-service"),
				testserver.WithNamespace(namespaceOutside),
			)).
			Install(testserver.Install(
				testserver.WithName("not-accessible-es"),
				testserver.WithNamespace(namespaceOutside),
			)).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		// Zone Kube2
		kubeServiceYAML := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshService
metadata:
  name: other-zone-test-server
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
    kuma.io/managed-by: k8s-controller
    k8s.kuma.io/is-headless-service: "false"
spec:
  selector:
    dataplaneTags:
      app: other-zone-test-server
      k8s.kuma.io/namespace: %s
  ports:
  - port: 80
    name: main
    targetPort: main
    appProtocol: http
`, namespace, meshName, namespace)
		kubeServiceNotAccessibleYAML := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshService
metadata:
  name: other-zone-not-accessible
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
    kuma.io/managed-by: k8s-controller
    k8s.kuma.io/is-headless-service: "false"
spec:
  selector:
    dataplaneTags:
      app: other-zone-not-accessible
      k8s.kuma.io/namespace: %s
  ports:
  - port: 80
    name: main
    targetPort: main
    appProtocol: http
`, namespace, meshName, namespace)
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("other-zone-test-server"),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "other-zone-test-server"),
			)).
			Install(testserver.Install(
				testserver.WithName("other-zone-not-accessible"),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "other-zone-not-accessible"),
			)).
			Install(YamlK8s(kubeServiceYAML)).
			Install(YamlK8s(kubeServiceNotAccessibleYAML)).
			Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespaceOutside)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to connect to reachable backends", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectFailure(
				multizone.KubeZone1, "client-server", "first-test-server.mesh",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			_, err := client.CollectFailure(
				multizone.KubeZone1, "client-server", "external-service.mesh",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			_, err := client.CollectFailure(
				multizone.KubeZone1, "client-server", "other-zone-test-server.global.mmzsreachable",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})

	It("should not connect to non reachable service", func() {
		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable services via Kuma DNS
			response, err := client.CollectFailure(
				multizone.KubeZone1, "client-server", "second-test-server.mesh",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			// then it fails because Kuma DP has no such DNS
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Equal(6))
		}, "5s", "100ms", MustPassRepeatedly(3)).Should(Succeed())

		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable service via Kubernetes DNS
			response, err := client.CollectFailure(
				multizone.KubeZone1, "client-server", "second-test-server.reachable-backends.svc.cluster.local",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			// then it fails because we don't encrypt traffic to unknown destination in the mesh
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(52), Equal(56)))
		}, "5s", "100ms", MustPassRepeatedly(3)).Should(Succeed())

		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable services via Kuma DNS
			response, err := client.CollectFailure(
				multizone.KubeZone1, "client-server", "not-accessible-es.mesh",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			// then it fails because Kuma DP has no such DNS
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Equal(6))
		}, "5s", "100ms", MustPassRepeatedly(3)).Should(Succeed())

		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable service via Kubernetes DNS
			response, err := client.CollectFailure(
				multizone.KubeZone1, "client-server", "not-accessible-es.reachable-backends-non-mesh.svc.cluster.local",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			// then it fails because we don't encrypt traffic to unknown destination in the mesh
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(52), Equal(56)))
		}, "5s", "100ms", MustPassRepeatedly(3)).Should(Succeed())

		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable mesh multizone service via Kuma DNS
			response, err := client.CollectFailure(
				multizone.KubeZone1, "client-server", "other-zone-not-accessible.global.mmzsreachable",
				client.FromKubernetesPod(namespace, "client-server"),
			)
			// then it fails because we don't encrypt traffic to unknown destination in the mesh
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(6)))
		}, "5s", "100ms", MustPassRepeatedly(3)).Should(Succeed())

		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable services via Kuma DNS
			response, err := client.CollectFailure(
				multizone.KubeZone1, "client-server-no-access", "second-test-server.mesh",
				client.FromKubernetesPod(namespace, "client-server-no-access"),
			)
			// then it fails because Kuma DP has no such DNS
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Equal(6))
		}, "5s", "100ms", MustPassRepeatedly(3)).Should(Succeed())

		Consistently(func(g Gomega) {
			// when trying to connect to non-reachable service via Kubernetes DNS
			response, err := client.CollectFailure(
				multizone.KubeZone1, "client-server-no-access", "second-test-server.reachable-backends.svc.cluster.local",
				client.FromKubernetesPod(namespace, "client-server-no-access"),
			)
			// then it fails because we don't encrypt traffic to unknown destination in the mesh
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(52), Equal(56)))
		}, "5s", "100ms", MustPassRepeatedly(3)).Should(Succeed())
	})
}
