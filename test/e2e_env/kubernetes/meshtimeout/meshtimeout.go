package meshtimeout

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshTimeout() {
	DescribeTableSubtree("with meshServices mode", func(mode mesh_proto.Mesh_MeshServices_Mode) {
		mesh := fmt.Sprintf("meshtimeout-ms-%s", strings.ToLower(mode.String()))
		namespace := fmt.Sprintf("%s-namespace", mesh)
		testServerURL := fmt.Sprintf("test-server.%s.svc:80", namespace)
		testServerSecondaryInboundUrl := fmt.Sprintf("test-server.%s.svc:8080", namespace)

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(Yaml(builders.Mesh().
					WithName(mesh).
					WithoutInitialPolicies().
					WithMeshServicesEnabled(mode))).
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(Parallel(
					democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(mesh)),
					testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(namespace)),
				)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEachFailure(func() {
			DebugKube(kubernetes.Cluster, mesh, namespace)
		})

		E2EAfterAll(func() {
			Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
			Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
		})

		DescribeTable("should add timeouts", FlakeAttempts(3), func(timeoutConfig string) {
			// Delete all retries and timeouts policy
			Expect(DeleteMeshResources(kubernetes.Cluster, mesh,
				meshtimeout_api.MeshTimeoutResourceTypeDescriptor,
				meshretry_api.MeshRetryResourceTypeDescriptor,
			)).To(Succeed())

			Eventually(func(g Gomega) {
				start := time.Now()
				g.Expect(client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", testServerURL,
					client.FromKubernetesPod(namespace, "demo-client"),
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10),
				)).Should(HaveField("Instance", ContainSubstring("test-server")))
				g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
			}, "30s", "1s").Should(Succeed())

			// when
			Expect(YamlK8s(timeoutConfig)(kubernetes.Cluster)).To(Succeed())

			// then
			Eventually(func(g Gomega) {
				g.Expect(client.CollectFailure(
					kubernetes.Cluster, "demo-client", testServerURL,
					client.FromKubernetesPod(namespace, "demo-client"),
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10), // we don't want 'curl' to return early
				)).Should(HaveField("ResponseCode", 504))
			}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
		},
			Entry("outbound", fmt.Sprintf(`
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
			Entry("inbound", fmt.Sprintf(`
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
			Entry("outbound dataplane kind", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: demo-client
  to:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, Config.KumaNamespace, mesh)),
			Entry("consumer policy", fmt.Sprintf(`
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
			func() []TableEntry { // Some tests don't run with all modes
				out := []TableEntry{}
				if mode == mesh_proto.Mesh_MeshServices_Exclusive { // These tests are only valid when using MeshService
					out = append(out, Entry("producer policy", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
        namespace: %s
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, namespace, mesh, namespace)))
				}
				return out
			}(),
		)

		It("should configure timeout for single inbound", func() {
			policy := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: test-server
    sectionName: secondary
  from:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, Config.KumaNamespace, mesh)

			// Delete all retries and timeouts policy
			Expect(DeleteMeshResources(kubernetes.Cluster, mesh,
				meshtimeout_api.MeshTimeoutResourceTypeDescriptor,
				meshretry_api.MeshRetryResourceTypeDescriptor,
			)).To(Succeed())

			// main inbound
			Eventually(func(g Gomega) {
				start := time.Now()
				g.Expect(client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", testServerURL,
					client.FromKubernetesPod(namespace, "demo-client"),
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10),
				)).Should(HaveField("Instance", ContainSubstring("test-server")))
				g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
			}, "30s", "1s").Should(Succeed())

			// secondary inbound
			Eventually(func(g Gomega) {
				start := time.Now()
				g.Expect(client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", testServerSecondaryInboundUrl,
					client.FromKubernetesPod(namespace, "demo-client"),
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10),
				)).Should(HaveField("Instance", ContainSubstring("test-server")))
				g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
			}, "30s", "1s").Should(Succeed())

			// when
			Expect(YamlK8s(policy)(kubernetes.Cluster)).To(Succeed())

			// then
			// main inbound
			Eventually(func(g Gomega) {
				start := time.Now()
				g.Expect(client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", testServerURL,
					client.FromKubernetesPod(namespace, "demo-client"),
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10),
				)).Should(HaveField("Instance", ContainSubstring("test-server")))
				g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
			}, "30s", "1s", MustPassRepeatedly(5)).Should(Succeed())

			// secondary inbound
			Eventually(func(g Gomega) {
				g.Expect(client.CollectFailure(
					kubernetes.Cluster, "demo-client", testServerSecondaryInboundUrl,
					client.FromKubernetesPod(namespace, "demo-client"),
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10), // we don't want 'curl' to return early
				)).Should(HaveField("ResponseCode", 504))
			}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
		})
	},
		Entry("Disabled", mesh_proto.Mesh_MeshServices_Disabled),
		Entry("Exclusive", mesh_proto.Mesh_MeshServices_Exclusive),
	)
}
