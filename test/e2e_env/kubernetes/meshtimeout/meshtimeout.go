package meshtimeout

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshTimeout() {
	namespace := "meshtimeout-namespace"
	mesh := "meshtimeout"

	kubeMeshServiceYAML := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshService
metadata:
  name: test-server
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
    kuma.io/managed-by: k8s-controller
    k8s.kuma.io/is-headless-service: "false"
spec:
  selector:
    dataplaneTags:
      app: test-server
      k8s.kuma.io/namespace: %s
  ports:
  - port: 80
    name: main
    targetPort: main
    appProtocol: http
`, namespace, mesh, namespace)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(mesh))).
			Install(testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(namespace))).
			Install(YamlK8s(kubeMeshServiceYAML)).
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: HostnameGenerator
metadata:
  name: timeout-connectivity
  namespace: %s
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.{{ .Zone }}.meshtimeout'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: zone
        kuma.io/managed-by: k8s-controller
`, Config.KumaNamespace))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshtimeout policy
		Expect(DeleteMeshPolicyOrError(
			kubernetes.Cluster,
			v1alpha1.MeshTimeoutResourceTypeDescriptor,
			fmt.Sprintf("mesh-timeout-all-%s", mesh),
		)).To(Succeed())

		// Delete the default meshretry policy
		Expect(DeleteMeshPolicyOrError(
			kubernetes.Cluster,
			meshretry_api.MeshRetryResourceTypeDescriptor,
			fmt.Sprintf("mesh-retry-all-%s", mesh),
		)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, namespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	DescribeTable("should add timeouts", FlakeAttempts(3), func(timeoutConfig string) {
		// given no MeshTimeout
		mts, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshtimeouts", mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(mts).To(
			Or(
				HaveExactElements(Equal("mesh-gateways-timeout-all-meshtimeout.kuma-system")),
				BeEmpty(),
			),
		)

		Eventually(func(g Gomega) {
			start := time.Now()
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithHeader("x-set-response-delay-ms", "5000"),
				client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
		}, "30s", "1s").Should(Succeed())

		// when
		Expect(YamlK8s(timeoutConfig)(kubernetes.Cluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithHeader("x-set-response-delay-ms", "5000"),
				client.WithMaxTime(10), // we don't want 'curl' to return early
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())

		Expect(DeleteYamlK8s(timeoutConfig)(kubernetes.Cluster)).To(Succeed())
	},
		Entry("outbound timeout", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, Config.KumaNamespace, mesh)),
		Entry("inbound timeout", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, Config.KumaNamespace, mesh)),
		Entry("consumer MeshTimeout policy", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, namespace, mesh)),
	)

	It("should target real MeshService resource", func() {
		// given no MeshTimeout
		mts, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshtimeouts", mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(mts).To(
			Or(
				HaveExactElements(Equal("mesh-gateways-timeout-all-meshtimeout.kuma-system")),
				BeEmpty(),
			),
		)

		Eventually(func(g Gomega) {
			start := time.Now()
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", fmt.Sprintf("test-server.%s.default.meshtimeout", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithHeader("x-set-response-delay-ms", "5000"),
				client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
		}, "30s", "1s").Should(Succeed())

		timeoutConfig := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
        namespace: %s
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, namespace, mesh, namespace)

		// when
		Expect(YamlK8s(timeoutConfig)(kubernetes.Cluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", fmt.Sprintf("test-server.%s.default.meshtimeout", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithHeader("x-set-response-delay-ms", "5000"),
				client.WithMaxTime(10), // we don't want 'curl' to return early
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())

		Expect(DeleteYamlK8s(timeoutConfig)(kubernetes.Cluster)).To(Succeed())
	})
}
