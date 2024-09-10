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

func MeshServicesWithReachableBackendsOption() {
	meshName := "mesh-service-reachable-backends"
	namespace := "mesh-service-reachable-backends"
	reachableBackends := `
      refs:
      - kind: MeshService
        labels:
          kuma.io/display-name: other-zone-test-server
      - kind: MeshService
        name: local-test-server
`

	mesh := fmt.Sprintf(`
type: Mesh
name: "%s"
meshServices:
  enabled: ReachableBackends
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
networking:
  outbound:
    passthrough: false
routing:
  zoneEgress: true
`, meshName)

	BeforeAll(func() {
		// Global
		err := NewClusterSetup().
			Install(YamlUniversal(mesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		// Zone Kube1
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("client-without-reachable"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Install(testserver.Install(
				testserver.WithName("client-with-reachable-backends-only"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithReachableBackends(reachableBackends),
			)).
			Install(testserver.Install(
				testserver.WithName("local-test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: HostnameGenerator
metadata:
  name: local-k8s-mesh-services
  namespace: %s
  labels:
    kuma.io/origin: zone
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.svc.{{ .Zone }}.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: zone
        kuma.io/env: kubernetes
`, Config.KumaNamespace))).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("other-zone-test-server"),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "other-zone-test-server"),
			)).
			Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to connect to k8s VIP, using legacy cluster, if reachable backends isn't set", func() {
		Eventually(func(g Gomega) {
			// when
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1, "client-without-reachable", "local-test-server.mesh-service-reachable-backends.svc.kuma-1.mesh.local",
				client.FromKubernetesPod(namespace, "client-without-reachable"),
			)
			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(HavePrefix("local-test-server"))
		}, "30s", "500ms", MustPassRepeatedly(10)).Should(Succeed())

		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(multizone.KubeZone1, "client-without-reachable", namespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := multizone.KubeZone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod+"."+namespace, "--type=clusters", fmt.Sprintf("--mesh=%s", meshName))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring(fmt.Sprintf("local-test-server_%s_svc_80", namespace)))
			g.Expect(stdout).ToNot(ContainSubstring("_msvc_"))
		}, "10s", "1s").Should(Succeed())
	})

	It("should not be able to connect to cross-zone MeshService if reachable backends isn't set", func() {
		Consistently(func(g Gomega) {
			// when
			resp, err := client.CollectFailure(
				multizone.KubeZone1, "client-without-reachable", "other-zone-test-server.mesh-service-reachable-backends.svc.kuma-2.mesh.local",
				client.FromKubernetesPod(namespace, "client-without-reachable"),
			)
			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Exitcode).To(Or(Equal(6), Equal(28)))
		}, "15s", "500ms", MustPassRepeatedly(5)).Should(Succeed())
	})

	It("should be able to connect if reachable backends is set", func() {
		Eventually(func(g Gomega) {
			// when
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1, "client-with-reachable-backends-only", "local-test-server.mesh-service-reachable-backends.svc.kuma-1.mesh.local",
				client.FromKubernetesPod(namespace, "client-with-reachable-backends-only"),
			)
			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(HavePrefix("local-test-server"))

			pod, err := PodNameOfApp(multizone.KubeZone1, "client-with-reachable-backends-only", namespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := multizone.KubeZone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod+"."+namespace, "--type=clusters", fmt.Sprintf("--mesh=%s", meshName))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring(fmt.Sprintf("%s_local-test-server_%s_kuma-1_msvc_80", meshName, namespace)))
		}, "10s", "500ms", MustPassRepeatedly(10)).Should(Succeed())

		Eventually(func(g Gomega) {
			// when
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1, "client-with-reachable-backends-only", "other-zone-test-server.mesh-service-reachable-backends.svc.kuma-2.mesh.local",
				client.FromKubernetesPod(namespace, "client-with-reachable-backends-only"),
			)
			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(HavePrefix("other-zone-test-server"))
		}, "10s", "500ms", MustPassRepeatedly(10)).Should(Succeed())
	})
}
